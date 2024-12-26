package main

import (
	"log/slog"
	"os"
	"os/signal"
	"proc/internal/app"
	"proc/internal/config"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger()
	log.Info("starting app")

	application := app.NewApp(log, cfg)
	application.StartNewApp()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("app started")
	<-done
	log.Info("app stopped")
}

func setupLogger() *slog.Logger {
	var logger *slog.Logger

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return logger
}
