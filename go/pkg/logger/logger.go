package logger

import (
	"log/slog"
	"os"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/config"
)

func NewLogger(
	logConfig *config.LogConfig,
) *slog.Logger {
	var level slog.LevelVar
	if err := level.UnmarshalText([]byte(logConfig.Level)); err != nil {
		panic(failure.Translate(err, appError.ErrConfigFileSyntaxError, failure.Messagef("invalid log level: %s", logConfig.Level)))
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: &level}))

	return logger
}
