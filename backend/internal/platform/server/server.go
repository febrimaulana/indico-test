// Package server wraps http.Server with timeouts and graceful shutdown.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Timeouts configures the HTTP server timeouts.
type Timeouts struct {
	ReadHeader time.Duration
	Read       time.Duration
	Write      time.Duration
	Idle       time.Duration
}

// Server runs an HTTP server until the context is canceled, then drains
// in-flight requests before returning.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration

	// listen and shutdown are seams so tests can exercise every branch
	// without opening real sockets.
	listen   func() error
	shutdown func(ctx context.Context) error
}

// New builds a Server with sane production defaults.
func New(handler http.Handler, addr string, timeouts Timeouts, shutdownTimeout time.Duration) *Server {
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: timeouts.ReadHeader,
		ReadTimeout:       timeouts.Read,
		WriteTimeout:      timeouts.Write,
		IdleTimeout:       timeouts.Idle,
	}
	return &Server{
		httpServer:      httpServer,
		shutdownTimeout: shutdownTimeout,
		listen:          httpServer.ListenAndServe,
		shutdown:        httpServer.Shutdown,
	}
}

// Run serves HTTP until ctx is canceled. On cancellation it drains requests
// (bounded by the shutdown timeout) and waits for the listener to exit.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		err := s.listen()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		if err := s.shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("gagal menghentikan server dengan rapi: %w", err)
		}
		<-errCh
		return nil
	}
}
