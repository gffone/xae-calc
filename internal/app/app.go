package app

import (
	"log/slog"
	"proc/internal/config"
	"proc/internal/server"
	"proc/internal/storage"
)

type App struct {
	log *slog.Logger
	cfg *config.Config
}

func NewApp(log *slog.Logger, cfg *config.Config) *App {
	return &App{log: log, cfg: cfg}
}

func (a *App) StartNewApp() {
	a.log.Info("create storage")
	strg := storage.NewStorage(a.cfg.StoragePath, a.cfg.WorksheetName)
	a.log.Info("storage created")

	a.log.Info("start new server")
	srv := server.NewServer(a.log, a.cfg.StaticDir, a.cfg.Url, a.cfg.Port, strg)
	srv.StartNewServer()
	a.log.Info("server started")
}
