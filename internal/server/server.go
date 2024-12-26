package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"proc/internal/server/handler"
	"proc/internal/storage"
)

type Server struct {
	log       *slog.Logger
	staticDir string
	url       string
	port      string
	strg      *storage.Storage
}

func NewServer(log *slog.Logger, staticDir string, url string, port string, strg *storage.Storage) *Server {
	return &Server{log: log, staticDir: staticDir, url: url, port: port, strg: strg}
}

func (s *Server) StartNewServer() {
	err := s.initializeFileServer()
	if err != nil {
		s.log.Error(err.Error())
	}

	err = s.initializeHandlerFunctions()
	if err != nil {
		s.log.Error(err.Error())
	}

	s.listenAndServe()
}

func (s *Server) initializeFileServer() error {
	s.log.Info("initializing handler with FileServer")

	if err := validateFolder(s.staticDir); err != nil {
		return errors.New("couldn't validate static folder, error: " + err.Error())
	}

	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir(s.staticDir))))
	s.log.Info("handler has been initialized")
	return nil
}

func (s *Server) initializeHandlerFunctions() error {
	s.log.Info("initializing the handler functions...")

	if err := validateFolder(s.staticDir); err != nil {
		return errors.New("couldn't validate static folder, error: " + err.Error())
	}

	handler.HandlerFunctions(s.log, s.strg)

	s.log.Info("handler functions have been initialized")
	return nil
}

func (s *Server) listenAndServe() {
	serverUrl := s.url + ":" + s.port

	s.log.Info(fmt.Sprintf("listening on %s...\n", serverUrl))

	go func() {
		if err := http.ListenAndServe(serverUrl, nil); err != nil {
			panic("failed to start server")
		}
	}()
}

func validateFolder(folderPath string) error {
	fileInfo, err := os.Stat(folderPath)

	if err != nil {
		return err
	}

	if fileInfo != nil && !fileInfo.IsDir() {
		return fmt.Errorf("%s : is not a directory", folderPath)
	}

	return nil
}
