package authn

import "context"

type Authenticator interface {
	Authenticate(context.Context) (string, error)
}

type NoopAuthenticator struct{}

func (a *NoopAuthenticator) Authenticate(ctx context.Context) (string, error) {
	return "", nil
}
