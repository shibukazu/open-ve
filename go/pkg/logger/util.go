package logger

import (
	"fmt"
	"log/slog"

	"github.com/morikuni/failure/v2"
)

func LogError(logger *slog.Logger, err error) {
	logger.Error(string(failure.MessageOf(err)), slog.Any("code", failure.CodeOf(err)), slog.String("details", fmt.Sprintf("%+v", err)))
}
