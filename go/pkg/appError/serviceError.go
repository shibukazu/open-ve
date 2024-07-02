package appError

import (
	"github.com/morikuni/failure/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
const (
	ErrServerStartFailed = "ServerStartFailed"

	ErrYAMLSyntaxError = "YAMLSyntaxError"
	ErrDSLSyntaxError = "DSLSyntaxError"
	ErrDSLNotFound = "DSLNotFound"
	ErrRedisOperationFailed = "RedisOperationFailed"
	ErrCELSyantaxError = "CELSyantaxError"

	ErrRequestParameterInvalid = "RequestParameterInvalid"

	ErrValidateServiceIDNotFound = "ValidateServiceIDNotFound"
	
	ErrDSLServiceDSLSyntaxError = "DSLServiceDSLSyntaxError"
)

func ToGRPCError(err error) error {
	var code codes.Code
	switch failure.CodeOf(err) {
	case ErrServerStartFailed:
		code = codes.Internal
	case ErrYAMLSyntaxError:
		code = codes.InvalidArgument
	case ErrDSLSyntaxError:
		code = codes.InvalidArgument
	case ErrRedisOperationFailed:
		code = codes.Internal
	case ErrCELSyantaxError:
		code = codes.InvalidArgument
	case ErrRequestParameterInvalid:
		code = codes.InvalidArgument
	case ErrValidateServiceIDNotFound:
		code = codes.NotFound
	case ErrDSLServiceDSLSyntaxError:
		code = codes.InvalidArgument
	default:
		code = codes.Unknown
	}
	return status.Error(code, getMessage(err))
}

func getMessage(err error) string {
	msg := failure.MessageOf(err)
	if msg != "" {
		return string(msg)
	}
	return "Error"
}