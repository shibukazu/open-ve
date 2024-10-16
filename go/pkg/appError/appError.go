package appError

import (
	"fmt"

	"github.com/morikuni/failure/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrConfigError             = "ConfigError"
	ErrDSLSyntaxError          = "DSLSyntaxError"
	ErrDSLGenerationFailed     = "DSLGenerationFailed"
	ErrStoreOperationFailed    = "StoreOperationFailed"
	ErrRequestParameterInvalid = "RequestParameterInvalid"
	ErrServerError             = "ServerError"
	ErrRequestForwardFailed    = "RequestForwardError"
	ErrAuthenticationFailed    = "AuthenticationFailed"
)

func ToGRPCError(err error) error {
	var code codes.Code
	switch failure.CodeOf(err) {
	case ErrConfigError:
		code = codes.InvalidArgument
	case ErrDSLSyntaxError:
		code = codes.InvalidArgument
	case ErrDSLGenerationFailed:
		code = codes.Internal
	case ErrStoreOperationFailed:
		code = codes.Internal
	case ErrRequestParameterInvalid:
		code = codes.InvalidArgument
	case ErrAuthenticationFailed:
		code = codes.Unauthenticated
	case ErrServerError:
		code = codes.Internal
	case ErrRequestForwardFailed:
		code = codes.Internal
	default:
		code = codes.Unknown
	}
	return status.Error(code, getGRPCErrorMessage(err))
}

func getGRPCErrorMessage(err error) string {
	code := failure.CodeOf(err)
	message := failure.MessageOf(err)
	cause := failure.CauseOf(err)
	ret := fmt.Sprintf("code: %s, message: %s, cause: %s", code, message, cause)

	return ret
}
