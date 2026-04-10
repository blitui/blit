package blit_test

import (
	"context"
	"net"
	"testing"
	"time"

	blit "github.com/blitui/blit"
	"golang.org/x/crypto/ssh"
)

// serveMinimalApp starts a Serve with a minimal App on addr, returning when
// ctx is cancelled.
func serveMinimalApp(ctx context.Context, addr string) error {
	return blit.Serve(ctx, blit.ServeConfig{
		Addr: addr,
		Factory: func() *blit.App {
			return blit.NewApp()
		},
	})
}

// TestServe_integration dials an ephemeral Wish SSH server and verifies a PTY
// session is accepted, confirming the factory pattern is wired correctly.
func TestServe_integration(t *testing.T) {
	// Pick an ephemeral port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- serveMinimalApp(ctx, addr)
	}()

	// Poll until TCP port is accepting connections.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, dialErr := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if dialErr == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Connect via SSH — accept any host key (test environment only).
	clientCfg := &ssh.ClientConfig{
		User:            "testuser",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // test only
		Timeout:         3 * time.Second,
	}
	client, sshErr := ssh.Dial("tcp", addr, clientCfg)
	if sshErr != nil {
		cancel()
		t.Fatalf("ssh.Dial: %v", sshErr)
	}
	defer client.Close()

	sess, sessErr := client.NewSession()
	if sessErr != nil {
		cancel()
		t.Fatalf("NewSession: %v", sessErr)
	}
	defer sess.Close()

	// Request a PTY — required by activeterm middleware.
	if ptyErr := sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{}); ptyErr != nil {
		cancel()
		t.Fatalf("RequestPty: %v", ptyErr)
	}
	if shellErr := sess.Shell(); shellErr != nil {
		cancel()
		t.Fatalf("Shell: %v", shellErr)
	}

	// Shut the server down and confirm it exits cleanly.
	cancel()
	select {
	case err := <-srvErr:
		if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
			t.Errorf("Serve returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("server did not shut down in time")
	}
}
