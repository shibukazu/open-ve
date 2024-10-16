package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/morikuni/failure/v2"
	"github.com/rs/cors"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	"github.com/shibukazu/open-ve/go/pkg/slave"
	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbSlave "github.com/shibukazu/open-ve/go/proto/slave/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
)

type Gateway struct {
	mode         string
	httpConfig   *config.HttpConfig
	gRPCConfig   *config.GRPCConfig
	logger       *slog.Logger
	dslReader    *reader.DSLReader
	slaveManager *slave.SlaveManager
	server       *http.Server
}

func NewGateway(
	mode string,
	httpConfig *config.HttpConfig,
	gRPCConfig *config.GRPCConfig,
	logger *slog.Logger,
	dslReader *reader.DSLReader,
	slaveManager *slave.SlaveManager,
) *Gateway {
	return &Gateway{
		mode:         mode,
		httpConfig:   httpConfig,
		gRPCConfig:   gRPCConfig,
		logger:       logger,
		dslReader:    dslReader,
		slaveManager: slaveManager,
	}
}

func (g *Gateway) Run(ctx context.Context, wg *sync.WaitGroup) {
	dialOpts := []grpc.DialOption{}

	if g.gRPCConfig.TLS.Enabled {
		if g.gRPCConfig.TLS.CertPath == "" || g.gRPCConfig.TLS.KeyPath == "" {
			panic(failure.New(appError.ErrServerError, failure.Message("certPath and keyPath must be set")))
		}
		creds, err := credentials.NewClientTLSFromFile(g.gRPCConfig.TLS.CertPath, "")
		if err != nil {
			panic(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to load TLS cert")))
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(":"+g.gRPCConfig.Port, dialOpts...)
	if err != nil {
		panic(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to dial gRPC server")))
	}
	defer conn.Close()

	runtime.DefaultContextTimeout = 10 * time.Second
	muxOpts := []runtime.ServeMuxOption{
		runtime.WithHealthzEndpoint(pbHealth.NewHealthClient(conn)),
	}
	grpcGateway := runtime.NewServeMux(muxOpts...)

	if err := pbValidate.RegisterValidateServiceHandlerFromEndpoint(ctx, grpcGateway, ":"+g.gRPCConfig.Port, dialOpts); err != nil {
		panic(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to register validate service on gateway")))
	}

	if err := pbDSL.RegisterDSLServiceHandlerFromEndpoint(ctx, grpcGateway, ":"+g.gRPCConfig.Port, dialOpts); err != nil {
		panic(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to register dsl service on gateway")))
	}

	if g.mode == "master" {
		if err := pbSlave.RegisterSlaveServiceHandlerFromEndpoint(ctx, grpcGateway, ":"+g.gRPCConfig.Port, dialOpts); err != nil {
			panic(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to register slave service on gateway")))
		}
	}

	withMiddleware := g.forwardCheckRequestMiddleware(g.validateRequestTypeConvertMiddleware(grpcGateway))

	withCors := cors.New(cors.Options{
		AllowedOrigins:   g.httpConfig.CORSAllowedOrigins,
		AllowedHeaders:   g.httpConfig.CORSAllowedHeaders,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(withMiddleware)

	g.server = &http.Server{
		Addr:    ":" + g.httpConfig.Port,
		Handler: withCors,
	}

	go func() {
		if g.httpConfig.TLS.Enabled {
			if g.httpConfig.TLS.CertPath == "" || g.httpConfig.TLS.KeyPath == "" {
				panic(failure.New(appError.ErrServerError, failure.Messagef("TLS certPath and keyPath must be specified")))
			}
			if err := g.server.ListenAndServeTLS(g.httpConfig.TLS.CertPath, g.httpConfig.TLS.KeyPath); err != nil && err != http.ErrServerClosed {
				panic(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to start gateway server with TLS")))
			}
		} else {
			if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				panic(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to start gateway server")))
			}
		}
	}()

	if g.httpConfig.TLS.Enabled {
		g.logger.Info("🔒 gateway server: TLS is enabled")
	}
	g.logger.Info("🟢 gateway server: started")

	// graceful shutdown
	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	g.shutdown(ctxShutDown)
	wg.Done()
}

func (g *Gateway) shutdown(ctx context.Context) {
	if err := g.server.Shutdown(ctx); err != nil {
		panic(failure.Translate(err, appError.ErrServerError, failure.Message("failed to shutdown gateway server")))
	}
	g.logger.Info("🛑 gateway server is stopped")
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
}

func (rec *responseRecorder) Write(b []byte) (int, error) {
	return rec.body.Write(b)
}

func (g *Gateway) forwardCheckRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if g.mode == "master" && r.URL.Path == "/v1/check" && r.Method == "POST" {
			ctx := context.Background()
			modifiedRequestValidations := make([]interface{}, 0)
			validationResults := make([]interface{}, 0)

			var reqBody map[string]interface{}
			var resBody map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				err = failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to decode request body"))
				http.Error(w, err.Error(), http.StatusBadRequest)
				logger.LogError(g.logger, err)
				return
			}

			validations, ok := reqBody["validations"].([]interface{})
			if !ok {
				err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validations field is invalid"))
				http.Error(w, err.Error(), http.StatusBadRequest)
				logger.LogError(g.logger, err)
				return
			}

			dslFound := false
			dsl, err := g.dslReader.Read(ctx)
			if err == nil {
				dslFound = true
			}

			ch := make(chan []interface{})
			errCh := make(chan error)
			numForwarded := 0
			for _, validation := range validations {
				validation, ok := validation.(map[string]interface{})
				if !ok {
					err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validation field is invalid"))
					http.Error(w, err.Error(), http.StatusBadRequest)
					logger.LogError(g.logger, err)
					return
				}
				id, ok := validation["id"].(string)
				if !ok {
					err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("id field is invalid"))
					http.Error(w, err.Error(), http.StatusBadRequest)
					logger.LogError(g.logger, err)
					return
				}

				// Check if the request forward is needed
				isForwardNeed := true
				if dslFound {
					for _, validation := range dsl.Validations {
						if validation.ID == id {
							isForwardNeed = false
							break
						}
					}
				}

				if isForwardNeed {
					numForwarded++
					go func(id string, ch chan []interface{}) {
						// Find the slave node that can handle validation ID
						slaveNode, err := g.slaveManager.FindSlave(id)
						if err != nil {
							errCh <- err
							return
						}

						var client *http.Client
						if slaveNode.TLSEnabled {
							transport := &http.Transport{
								TLSClientConfig: &tls.Config{},
							}
							client = &http.Client{Transport: transport}
						} else {
							client = &http.Client{}
						}
						client.Timeout = 5 * time.Second

						reqBody := map[string]interface{}{
							"validations": []interface{}{validation},
						}
						body, err := json.Marshal(reqBody)
						if err != nil {
							errCh <- failure.Translate(err, appError.ErrRequestForwardFailed, failure.Messagef("failed to marshal request body"))
							return
						}
						req, err := http.NewRequest("POST", slaveNode.Addr+"/v1/check", bytes.NewBuffer(body))
						if err != nil {
							errCh <- failure.Translate(err, appError.ErrRequestForwardFailed, failure.Messagef("failed to create forward equest"))
							return
						}
						req.Header.Set("Content-Type", "application/json")

						switch slaveNode.Authn.Method {
						case "preshared":
							req.Header.Set("Authorization", "Bearer "+slaveNode.Authn.Preshared.Key)
						}

						resp, err := client.Do(req)
						if err != nil {
							errCh <- failure.Translate(err, appError.ErrRequestForwardFailed, failure.Messagef("failed to forward request to slave id:%s", id))
							return
						}
						defer resp.Body.Close()

						if resp.StatusCode != http.StatusOK {
							errCh <- failure.New(appError.ErrRequestForwardFailed, failure.Messagef("failed to forward request to slave id:%s", id))
							return
						}

						var respBody map[string]interface{}
						if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
							errCh <- failure.Translate(err, appError.ErrRequestForwardFailed, failure.Messagef("failed to decode response body"))
							return
						}
						results, ok := respBody["results"].([]interface{})
						if !ok {
							errCh <- failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("results field is invalid"))
							return
						}
						ch <- results
						g.logger.Info(fmt.Sprintf("⚽️ Request (id:%s) Forwarded to Slave %s", id, slaveNode.Id))
					}(id, ch)
				} else {
					modifiedRequestValidations = append(modifiedRequestValidations, validation)
				}
			}

			for i := 0; i < numForwarded; i++ {
				select {
				case err := <-errCh:
					http.Error(w, err.Error(), http.StatusInternalServerError)
					logger.LogError(g.logger, err)
					return
				case results := <-ch:
					validationResults = append(validationResults, results...)
				case <-time.After(30 * time.Second):
					err := failure.New(appError.ErrRequestForwardFailed, failure.Message("request forward timeout"))
					http.Error(w, err.Error(), http.StatusInternalServerError)
					logger.LogError(g.logger, err)
				}
			}

			reqBody["validations"] = modifiedRequestValidations
			modifiedReqBody, err := json.Marshal(reqBody)
			if err != nil {
				err = failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to marshal modified request body"))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.LogError(g.logger, err)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(modifiedReqBody))
			r.ContentLength = int64(len(modifiedReqBody))

			rec := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
			}
			next.ServeHTTP(rec, r)

			// Concat the validation results
			if err := json.Unmarshal(rec.body.Bytes(), &resBody); err != nil {
				err = failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to decode response body"))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.LogError(g.logger, err)
				return
			}
			originalValidationResults, ok := resBody["results"].([]interface{})
			if !ok {
				err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("results field is invalid"))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.LogError(g.logger, err)
				return
			}
			resBody["results"] = append(originalValidationResults, validationResults...)
			resBodyJson, err := json.Marshal(resBody)
			if err != nil {
				err = failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to marshal response body"))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				logger.LogError(g.logger, err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", fmt.Sprint(len(resBodyJson)))
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(resBodyJson)
			if err != nil {
				g.logger.Error(failure.Translate(err, appError.ErrServerError, failure.Messagef("failed to write response")).Error())
			}
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (g *Gateway) validateRequestTypeConvertMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/check" && r.Method == "POST" {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				err = failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to decode request body"))
				http.Error(w, err.Error(), http.StatusBadRequest)
				logger.LogError(g.logger, err)
				return
			}

			validations, ok := body["validations"].([]interface{})
			if !ok {
				err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validations field is invalid"))
				http.Error(w, err.Error(), http.StatusBadRequest)
				logger.LogError(g.logger, err)
				return
			}

			for idx, validation := range validations {
				validation, ok := validation.(map[string]interface{})
				if !ok {
					err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validation field is invalid"))
					http.Error(w, err.Error(), http.StatusBadRequest)
					logger.LogError(g.logger, err)
					return
				}

				id, ok := validation["id"].(string)
				if !ok {
					err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("id field is invalid"))
					http.Error(w, err.Error(), http.StatusBadRequest)
					logger.LogError(g.logger, err)
					return
				}

				variables, ok := validation["variables"].(map[string]interface{})
				if !ok {
					err := failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("variables field is invalid"))
					http.Error(w, err.Error(), http.StatusBadRequest)
					logger.LogError(g.logger, err)
					return
				}

				variableNameToCELType, err := g.dslReader.GetVariableNameToCELType(context.Background(), id)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					logger.LogError(g.logger, err)
					return
				}

				convertedVariables := make(map[string]interface{}, len(variables))
				for key, value := range variables {
					celType := variableNameToCELType[key]
					convertedType, err := convertCELTypeToGoogleProtobufType(celType)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						logger.LogError(g.logger, err)
						return
					}
					variable := make(map[string]interface{}, 2)
					variable["@type"] = convertedType
					variable["value"] = value

					convertedVariables[key] = variable
				}
				validation["variables"] = convertedVariables

				validations[idx] = validation
			}
			body["validations"] = validations
			convertedBody, err := json.Marshal(body)
			if err != nil {
				err = failure.Translate(err, appError.ErrRequestParameterInvalid, failure.Messagef("failed to marshal request body"))
				http.Error(w, err.Error(), http.StatusBadRequest)
				logger.LogError(g.logger, err)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(convertedBody))
			r.ContentLength = int64(len(convertedBody))
		}
		next.ServeHTTP(w, r)
	})
}

func convertCELTypeToGoogleProtobufType(celType string) (string, error) {
	switch celType {
	case "int":
		return "type.googleapis.com/google.protobuf.Int64Value", nil
	case "uint":
		return "type.googleapis.com/google.protobuf.UInt64Value", nil
	case "double":
		return "type.googleapis.com/google.protobuf.DoubleValue", nil
	case "bool":
		return "type.googleapis.com/google.protobuf.BoolValue", nil
	case "string":
		return "type.googleapis.com/google.protobuf.StringValue", nil
	case "bytes":
		return "type.googleapis.com/google.protobuf.BytesValue", nil
	default:
		return "", failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("unsupported variable type: %s\nplease specify one of the following types: int, uint, double, bool, string, bytes", celType))
	}
}
