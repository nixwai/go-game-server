package bootstrap_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nixwai/go-game-server/app/bootstrap"
)

type fakeHTTPServer struct {
	listenErr   error
	release     chan struct{}
	shutdownErr error
}

func newFakeHTTPServer(listenErr, shutdownErr error) *fakeHTTPServer {
	return &fakeHTTPServer{
		listenErr:   listenErr,
		release:     make(chan struct{}),
		shutdownErr: shutdownErr,
	}
}

func (s *fakeHTTPServer) ListenAndServe() error {
	<-s.release
	return s.listenErr
}

func (s *fakeHTTPServer) Shutdown(context.Context) error {
	close(s.release)
	return s.shutdownErr
}

func TestRunHTTPServerReturnsUnexpectedListenError(t *testing.T) {
	server := newFakeHTTPServer(errors.New("listen failed"), nil)
	close(server.release)

	err := bootstrap.RunHTTPServer(context.Background(), server, time.Second)
	if err == nil || !strings.Contains(err.Error(), "listen failed") {
		t.Fatalf("expected listen error, got %v", err)
	}
}

func TestRunHTTPServerGracefullyShutsDownOnContextCancellation(t *testing.T) {
	server := newFakeHTTPServer(nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := bootstrap.RunHTTPServer(ctx, server, time.Second); err != nil {
		t.Fatalf("expected graceful shutdown, got %v", err)
	}
}

func TestRunHTTPServerReturnsShutdownError(t *testing.T) {
	server := newFakeHTTPServer(nil, errors.New("shutdown failed"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := bootstrap.RunHTTPServer(ctx, server, time.Second)
	if err == nil || !strings.Contains(err.Error(), "shutdown failed") {
		t.Fatalf("expected shutdown error, got %v", err)
	}
}
