package validatev1

import (
	"context"
	"log/slog"

	"github.com/shibukazu/open-ve/go/pkg/validator"
	pb "github.com/shibukazu/open-ve/go/proto/validate/v1"
)

type Service struct {
	pb.UnimplementedValidateServiceServer
	logger    *slog.Logger
	validator *validator.Validator
}

func NewService(ctx context.Context, logger *slog.Logger, validator *validator.Validator) *Service {
	return &Service{logger: logger, validator: validator}
}
