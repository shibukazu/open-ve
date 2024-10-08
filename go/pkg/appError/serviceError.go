package appError

import (
	"fmt"

	"github.com/morikuni/failure/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrYAMLSyntaxError      = "YAMLSyntaxError"
	ErrDSLSyntaxError       = "DSLSyntaxError"
	ErrRedisOperationFailed = "RedisOperationFailed"
	ErrCELSyantaxError      = "CELSyantaxError"

	ErrRequestParameterInvalid = "RequestParameterInvalid"

	ErrValidateServiceIDNotFound    = "ValidateServiceIDNotFound"
	ErrValidateServiceForwardFailed = "ValidateServiceForwardFailed"

	ErrDSLServiceDSLSyntaxError = "DSLServiceDSLSyntaxError"

	ErrAuthMissingToken = "AuthMissingToken"
	ErrAuthUnauthorized = "AuthUnauthorized"

	ErrInternalError = "Internal"
)

func ToGRPCError(err error) error {
	var code codes.Code
	switch failure.CodeOf(err) {
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
	case ErrAuthMissingToken:
		code = codes.Unauthenticated
	case ErrAuthUnauthorized:
		code = codes.Unauthenticated
	case ErrInternalError:
		code = codes.Internal
	default:
		code = codes.Unknown
	}
	return status.Error(code, getMessage(err))
}

func getMessage(err error) string {
	code := failure.CodeOf(err)
	cause := failure.CauseOf(err)
	additionalInfo := failure.MessageOf(err)
	detail := fmt.Sprintf("%+v\n", err)
	message := fmt.Sprintf("code: %s, cause: %s, additionalInfo: %s, detail: %s", code, cause, additionalInfo, detail)

	return message
}
