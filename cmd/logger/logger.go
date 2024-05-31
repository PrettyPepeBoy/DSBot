package logger

import (
	"DiscordBot/internal/lib/logger/prettySlog"
	"log/slog"
	"os"
)

const (
	environmentLocal = "local"
	environmentProd  = "prod"
)

func MustSetupLogger(environment string) *slog.Logger {
	return setupLogger(environment)
}

func setupLogger(environment string) *slog.Logger {
	var log *slog.Logger
	switch environment {
	case environmentLocal:
		log = setupPrettySlog()
	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := prettySlog.PrettyHandlersOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
