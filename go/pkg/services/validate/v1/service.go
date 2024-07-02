package validatev1

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/validator"
	pb "github.com/shibukazu/open-ve/go/proto/validate/v1"
)

type Service struct {
	pb.UnimplementedValidateServiceServer
	validator *validator.Validator
}

func NewService(ctx context.Context,validator *validator.Validator) *Service {
	return &Service{validator: validator}
}