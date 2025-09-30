package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	logger     *zap.Logger
	httpServer *http.Server
	onShutdown func() error
	close      func() error
}

type NetworkAddress struct {
	Host string
	Port int
}

const hostWildcard = "*"

func (addr NetworkAddress) String() string {
	host := addr.Host
	if host == hostWildcard {
		host = ""
	}
	return fmt.Sprintf("%s:%d", host, addr.Port)
}

func NewServer(logger *zap.Logger, addr NetworkAddress) *Server {
	return &Server{
		logger: logger,
		httpServer: &http.Server{
			Addr: addr.String(),
		},
	}
}

func (s *Server) Run() error {
	if err := s.httpServer.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}

func (s *Server) Address() string {
	return s.httpServer.Addr
}

func (s *Server) SetHandler(h http.Handler) {
	s.httpServer.Handler = h
}

func (s *Server) OnShutdown(fn func() error) {
	s.onShutdown = fn
}

func (s *Server) ShutdownHandler(osSignalEnabled bool, cancelFunc func() error) func() error {
	s.close = func() error {
		s.logger.Info("shutdown signal received")

		defer s.logger.Sync()

		// Create a context with a timeout for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Initiate graceful shutdown
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("HTTP server shutdown error", zap.Error(err))
		}
		s.logger.Info("HTTP server stopped")
		// Shutdown background worker goroutines
		if err := cancelFunc(); err != nil {
			s.logger.Error("shutdown cancellation callback function error", zap.Error(err))
		}
		if s.onShutdown != nil {
			if err := s.onShutdown(); err != nil {
				s.logger.Error("shutdown cleanup function error", zap.Error(err))
			}
		}
		s.logger.Info("shutdown callback executed")
		return nil
	}
	if !osSignalEnabled {
		return func() error {
			return nil
		}
	}
	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)

	// Subscribe to SIGINT (Ctrl+C) and SIGTERM (usually sent by container orchestrators)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	return func() error {
		<-quit

		return s.close()
	}
}

func (s *Server) Close() error {
	return s.close()
}
