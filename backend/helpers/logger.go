package helpers

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Log = slog.New(handler)
}
