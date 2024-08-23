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
	"github.com/shibukazu/open-ve/go/pkg/slave"
	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbSlave "github.com/shibukazu/open-ve/go/proto/slave/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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
	grpcGateway := runtime.NewServeMux()

	dialOpts := []grpc.DialOption{}

	if g.gRPCConfig.TLS.Enabled {
		if g.gRPCConfig.TLS.CertPath == "" || g.gRPCConfig.TLS.KeyPath == "" {
			panic(failure.New(appError.ErrServerStartFailed, failure.Message("certPath and keyPath must be set")))
		}
		creds, err := credentials.NewClientTLSFromFile(g.gRPCConfig.TLS.CertPath, "")
		if err != nil {
			panic(failure.Translate(err, appError.ErrServerStartFailed))
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err := pbValidate.RegisterValidateServiceHandlerFromEndpoint(ctx, grpcGateway, ":"+g.gRPCConfig.Port, dialOpts); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed, failure.Messagef("failed to register validate service on gateway")))
	}

	if err := pbDSL.RegisterDSLServiceHandlerFromEndpoint(ctx, grpcGateway, ":"+g.gRPCConfig.Port, dialOpts); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed, failure.Messagef("failed to register dsl service on gateway")))
	}

	if g.mode == "master" {
		if err := pbSlave.RegisterSlaveServiceHandlerFromEndpoint(ctx, grpcGateway, ":"+g.gRPCConfig.Port, dialOpts); err != nil {
			panic(failure.Translate(err, appError.ErrServerStartFailed, failure.Messagef("failed to register slave service on gateway")))
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
				panic(failure.New(appError.ErrServerStartFailed, failure.Messagef("TLS certPath and keyPath must be specified")))
			}
			if err := g.server.ListenAndServeTLS(g.httpConfig.TLS.CertPath, g.httpConfig.TLS.KeyPath); err != nil && err != http.ErrServerClosed {
				g.logger.Error(failure.Translate(err, appError.ErrServerInternalError).Error())
			}
		} else {
			if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				g.logger.Error(failure.Translate(err, appError.ErrServerInternalError).Error())
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
		g.logger.Error(failure.Translate(err, appError.ErrServerShutdownFailed).Error())
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
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusBadRequest)
				return
			}

			validations, ok := reqBody["validations"].([]interface{})
			if !ok {
				http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validations field is invalid")).Error(), http.StatusBadRequest)
				return
			}

			dslFound := false
			dsl, err := g.dslReader.Read(ctx)
			if err == nil {
				dslFound = true
			}
			// TODO: 各処理を並列化する
			for _, validation := range validations {
				validation, ok := validation.(map[string]interface{})
				if !ok {
					http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validation field is invalid")).Error(), http.StatusBadRequest)
					return
				}
				id, ok := validation["id"].(string)
				if !ok {
					http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("id field is invalid")).Error(), http.StatusBadRequest)
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
					// Find the slave node that can handle validation ID
					slaveNode, err := g.slaveManager.FindSlave(id)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
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
						http.Error(w, failure.Translate(err, appError.ErrValidateServiceForwardFailed).Error(), http.StatusInternalServerError)
						return
					}
					req, err := http.NewRequest("POST", slaveNode.Addr+"/v1/check", bytes.NewBuffer(body))
					if err != nil {
						http.Error(w, failure.Translate(err, appError.ErrValidateServiceForwardFailed).Error(), http.StatusInternalServerError)
						return
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := client.Do(req)
					if err != nil {
						http.Error(w, failure.Translate(err, appError.ErrValidateServiceForwardFailed).Error(), http.StatusInternalServerError)
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						http.Error(w, failure.New(appError.ErrValidateServiceForwardFailed, failure.Messagef("Failed to forward the validate request to slave: %d", resp.StatusCode)).Error(), http.StatusInternalServerError)
						return
					}

					var respBody map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
						http.Error(w, failure.Translate(err, appError.ErrValidateServiceForwardFailed).Error(), http.StatusInternalServerError)
						return
					}
					results, ok := respBody["results"].([]interface{})
					if !ok {
						http.Error(w, failure.New(appError.ErrValidateServiceForwardFailed, failure.Messagef("results field is invalid")).Error(), http.StatusInternalServerError)
						return
					}
					validationResults = append(validationResults, results...)

					g.logger.Info(fmt.Sprintf("⚽️ Request (id:%s) Forwarded to Slave %s", id, slaveNode.Id))
				} else {
					modifiedRequestValidations = append(modifiedRequestValidations, validation)
				}
			}

			reqBody["validations"] = modifiedRequestValidations
			modifiedReqBody, err := json.Marshal(reqBody)
			if err != nil {
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusInternalServerError)
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
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusInternalServerError)
				return
			}
			originalValidationResults, ok := resBody["results"].([]interface{})
			if !ok {
				http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("results field is invalid")).Error(), http.StatusInternalServerError)
				return
			}
			resBody["results"] = append(originalValidationResults, validationResults...)
			resBodyJson, err := json.Marshal(resBody)
			if err != nil {
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", fmt.Sprint(len(resBodyJson)))
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(resBodyJson)
			if err != nil {
				g.logger.Error(failure.Translate(err, appError.ErrServerInternalError).Error())
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
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusBadRequest)
				return
			}

			validations, ok := body["validations"].([]interface{})
			if !ok {
				http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validations field is invalid")).Error(), http.StatusBadRequest)
				return
			}

			for idx, validation := range validations {
				validation, ok := validation.(map[string]interface{})
				if !ok {
					http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("validation field is invalid")).Error(), http.StatusBadRequest)
					return
				}

				id, ok := validation["id"].(string)
				if !ok {
					http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("id field is invalid")).Error(), http.StatusBadRequest)
					return
				}

				variables, ok := validation["variables"].(map[string]interface{})
				if !ok {
					http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("variables field is invalid")).Error(), http.StatusBadRequest)
					return
				}

				variableNameToCELType, err := g.dslReader.GetVariableNameToCELType(context.Background(), id)
				if err != nil {
					http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusBadRequest)
					return
				}

				convertedVariables := make(map[string]interface{}, len(variables))
				for key, value := range variables {
					celType := variableNameToCELType[key]
					convertedType, err := convertCELTypeToGoogleProtobufType(celType)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
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
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusInternalServerError)
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
		return "", failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("Unsupported variable type: %s\nPlease specify one of the following types: int, uint, double, bool, string, bytes", celType))
	}
}
