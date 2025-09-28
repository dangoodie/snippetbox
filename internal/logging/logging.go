package logging

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func NewLogger() *slog.Logger {
	logger := slog.New(tint.NewHandler(os.Stderr, nil))

	return logger
}
