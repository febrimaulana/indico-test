package server

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestServer() *Server {
	return New(http.NewServeMux(), ":0", Timeouts{
		ReadHeader: time.Second,
		Read:       time.Second,
		Write:      time.Second,
		Idle:       time.Second,
	}, time.Second)
}

func TestRunListenerError(t *testing.T) {
	s := newTestServer()
	s.listen = func() error { return errors.New("listen gagal") }

	err := s.Run(context.Background())
	require.ErrorContains(t, err, "listen gagal")
}

func TestRunListenerClosed(t *testing.T) {
	s := newTestServer()
	s.listen = func() error { return http.ErrServerClosed }

	require.NoError(t, s.Run(context.Background()))
}

func TestRunGracefulShutdown(t *testing.T) {
	s := newTestServer()
	released := make(chan struct{})
	s.listen = func() error {
		<-released
		return nil
	}
	s.shutdown = func(context.Context) error {
		close(released)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, s.Run(ctx))
}

func TestRunShutdownError(t *testing.T) {
	s := newTestServer()
	s.listen = func() error {
		select {} // blocks forever; the test must not depend on it returning
	}
	s.shutdown = func(context.Context) error { return errors.New("shutdown gagal") }

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.Run(ctx)
	require.ErrorContains(t, err, "gagal menghentikan server dengan rapi")
}
