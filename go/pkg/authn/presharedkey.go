package authn

import (
	"context"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
)

type PresharedKeyAuthenticator struct {
	key string
}

func NewPresharedKeyAuthenticator(key string) *PresharedKeyAuthenticator {
	return &PresharedKeyAuthenticator{key: key}
}

func (a *PresharedKeyAuthenticator) Authenticate(ctx context.Context) (string, error) {
	authHeader, err := grpcauth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return "", failure.Translate(err, appError.ErrAuthenticationFailed, failure.Messagef("failed to get auth header"))
	}

	if authHeader != a.key {
		return "", failure.New(appError.ErrAuthenticationFailed, failure.Messagef("invalid key"))
	}

	return "", nil
}
