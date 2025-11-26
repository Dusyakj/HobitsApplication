package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"HobitsService/internal/logger"
)

// Server HTTP сервер для метрик и других эндпойнтов
type Server struct {
	server *http.Server
	mux    *http.ServeMux
}

// NewServer создает новый HTTP сервер
func NewServer(port int) *Server {
	mux := http.NewServeMux()

	// Регистрируем prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return &Server{
		server: httpServer,
		mux:    mux,
	}
}

// Start запускает HTTP сервер
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.server.Addr, err)
	}

	logger.Info("HTTP server started",
		zap.String("addr", s.server.Addr),
	)

	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

// Stop останавливает HTTP сервер
func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	logger.Info("HTTP server stopped", zap.String("addr", s.server.Addr))
	return nil
}

// AddHandler добавляет обработчик маршрута
func (s *Server) AddHandler(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// AddHandlerFunc добавляет функцию-обработчик маршрута
func (s *Server) AddHandlerFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
}
