// Package app provides the HTTP server lifecycle management.
package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

const (
	readTimeout     = 15 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 30 * time.Second
)

// Server wraps an *http.Server and owns its lifecycle.
type Server struct {
	httpServer *http.Server
}

// NewServer creates an HTTP server bound to addr using the provided router.
func NewServer(addr string, router *mux.Router) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}
}

// HTTPServer returns the underlying *http.Server. Exposed for testing.
func (s *Server) HTTPServer() *http.Server {
	return s.httpServer
}

// Start begins serving requests in a background goroutine and blocks until
// SIGINT or SIGTERM is received, then performs a graceful shutdown.
func (s *Server) Start() {
	go func() {
		log.Printf("Server listening on %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	s.shutdown()
	log.Println("Server exited")
}

func (s *Server) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
