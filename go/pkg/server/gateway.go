package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/morikuni/failure/v2"
	"github.com/rs/cors"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/config"
	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	httpConfig *config.HttpConfig
	gRPCConfig *config.GRPCConfig
	logger     *slog.Logger
}

func NewGateway(
	httpConfig *config.HttpConfig,
	gRPCConfig *config.GRPCConfig,
	logger *slog.Logger,
) *Gateway {
	return &Gateway{
		httpConfig: httpConfig,
		gRPCConfig: gRPCConfig,
		logger:     logger,
	}
}

func (g *Gateway) Run(ctx context.Context) {
	grpcGateway := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := pbValidate.RegisterValidateServiceHandlerFromEndpoint(ctx, grpcGateway, g.gRPCConfig.Addr, opts); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed, failure.Messagef("failed to register validate service on gateway")))
	}

	if err := pbDSL.RegisterDSLServiceHandlerFromEndpoint(ctx, grpcGateway, g.gRPCConfig.Addr, opts); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed, failure.Messagef("failed to register dsl service on gateway")))
	}

	withMiddleware := validateRequestTypeConvertMiddleware(grpcGateway)

	withCors := cors.New(cors.Options{
		AllowedOrigins:   g.httpConfig.CORSAllowedOrigins,
		AllowedHeaders:   g.httpConfig.CORSAllowedHeaders,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(withMiddleware)

	if err := http.ListenAndServe(g.httpConfig.Addr, withCors); err != nil {
		panic(err)
	}
}

func validateRequestTypeConvertMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/check" && r.Method == "POST" {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusBadRequest)
				return
			}

			variables, ok := body["variables"].(map[string]interface{})
			if !ok {
				http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("variables field is invalid")).Error(), http.StatusBadRequest)
				return
			}
			for key, value := range variables {
				variable, ok := value.(map[string]interface{})
				if !ok {
					http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("variable %s is invalid", key)).Error(), http.StatusBadRequest)
					return
				}
				variableType, ok := variable["type"].(string)
				if !ok {
					http.Error(w, failure.New(appError.ErrRequestParameterInvalid, failure.Messagef("variable %s is invalid", key)).Error(), http.StatusBadRequest)
					return
				}
				convertedType, err := convertCELTypeToGoogleProtobufType(variableType)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				delete(variable, "type")
				variable["@type"] = convertedType
				variables[key] = variable
			}
			body["variables"] = variables
			modifiedBody, err := json.Marshal(body)
			if err != nil {
				http.Error(w, failure.Translate(err, appError.ErrRequestParameterInvalid).Error(), http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(modifiedBody))
			r.ContentLength = int64(len(modifiedBody))
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
