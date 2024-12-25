package app

import (
	"log/slog"
	"proc/internal/config"
	"proc/internal/server"
	"proc/internal/storage"
)

func StartNewApp(log *slog.Logger, cfg *config.Config) {
	log.Info("create storage")
	strg := storage.NewStorage(cfg.StoragePath, cfg.WorksheetName)
	log.Info("storage created")

	log.Info("start new server")
	server.StartNewServer(log, cfg.StaticDir, cfg.Url, cfg.Port, strg)
	log.Info("server started")
}
