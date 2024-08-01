package server

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	httpConfig *config.HttpConfig
	gRPCConfig *config.GRPCConfig
	logger     *slog.Logger
	dslReader  *dsl.DSLReader
	server     *http.Server
}

func NewGateway(
	httpConfig *config.HttpConfig,
	gRPCConfig *config.GRPCConfig,
	logger *slog.Logger,
	dslReader *dsl.DSLReader,
) *Gateway {
	return &Gateway{
		httpConfig: httpConfig,
		gRPCConfig: gRPCConfig,
		logger:     logger,
		dslReader:  dslReader,
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

	if err := pbValidate.RegisterValidateServiceHandlerFromEndpoint(ctx, grpcGateway, g.gRPCConfig.Addr, dialOpts); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed, failure.Messagef("failed to register validate service on gateway")))
	}

	if err := pbDSL.RegisterDSLServiceHandlerFromEndpoint(ctx, grpcGateway, g.gRPCConfig.Addr, dialOpts); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed, failure.Messagef("failed to register dsl service on gateway")))
	}

	withMiddleware := g.validateRequestTypeConvertMiddleware(grpcGateway)

	withCors := cors.New(cors.Options{
		AllowedOrigins:   g.httpConfig.CORSAllowedOrigins,
		AllowedHeaders:   g.httpConfig.CORSAllowedHeaders,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(withMiddleware)

	g.server = &http.Server{
		Addr:    g.httpConfig.Addr,
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
