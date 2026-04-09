package tuikit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	btwish "github.com/charmbracelet/wish/bubbletea"
)

// ServeConfig configures the SSH server started by Serve.
type ServeConfig struct {
	// Addr is the host:port to listen on. Defaults to ":2222".
	Addr string

	// HostKeyPath is the path to the ED25519 host key file.
	// If empty, defaults to ~/.ssh/tuikit_host_ed25519 and is auto-generated
	// with 0600 permissions on first run.
	HostKeyPath string

	// Middleware is a list of additional Wish middleware to apply after the
	// built-in bubbletea and activeterm middlewares.
	Middleware []wish.Middleware

	// AuthorizedKeys is the path to an SSH authorized_keys file used to
	// restrict access. If empty, all connections are accepted (use only in
	// trusted networks).
	AuthorizedKeys string

	// Factory is called once per incoming SSH session to produce a fresh *App.
	// Each connection gets its own App instance so state is per-connection.
	Factory func() *App
}

// Serve hosts a tuikit App over SSH via Charm Wish. It blocks until ctx is
// cancelled or the server exits.
//
// Each inbound SSH session receives a fresh *App produced by cfg.Factory,
// ensuring per-connection state isolation.
//
//	tuikit.Serve(ctx, tuikit.ServeConfig{
//	    Addr:    ":2222",
//	    Factory: func() *tuikit.App { return tuikit.NewApp(...) },
//	})
func Serve(ctx context.Context, cfg ServeConfig) error {
	if cfg.Addr == "" {
		cfg.Addr = ":2222"
	}
	if cfg.Factory == nil {
		return fmt.Errorf("tuikit.Serve: Factory must not be nil")
	}

	hostKeyPath, err := resolveHostKeyPath(cfg.HostKeyPath)
	if err != nil {
		return fmt.Errorf("tuikit.Serve: host key: %w", err)
	}

	handler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		app := cfg.Factory()
		opts := []tea.ProgramOption{tea.WithAltScreen()}
		if app.model.mouseSupport {
			opts = append(opts, tea.WithMouseCellMotion())
		}
		return app.model, opts
	}

	mw := []wish.Middleware{
		btwish.Middleware(handler),
		activeterm.Middleware(),
	}
	mw = append(mw, cfg.Middleware...)

	serverOpts := []ssh.Option{
		wish.WithAddress(cfg.Addr),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMiddleware(mw...),
	}
	if cfg.AuthorizedKeys != "" {
		serverOpts = append(serverOpts, wish.WithAuthorizedKeys(cfg.AuthorizedKeys))
	}

	srv, err := wish.NewServer(serverOpts...)
	if err != nil {
		return fmt.Errorf("tuikit.Serve: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		if serveErr := srv.ListenAndServe(); serveErr != nil && serveErr != ssh.ErrServerClosed {
			errCh <- serveErr
		} else {
			errCh <- nil
		}
	}()

	select {
	case <-ctx.Done():
		// Try graceful shutdown first, but fall back to Close() if it
		// doesn't return within 2s — an active bubbletea session will
		// otherwise keep Shutdown blocked indefinitely.
		shutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		shutDone := make(chan struct{})
		go func() {
			_ = srv.Shutdown(shutCtx)
			close(shutDone)
		}()
		select {
		case <-shutDone:
		case <-shutCtx.Done():
			_ = srv.Close()
		}
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// resolveHostKeyPath returns the effective host key path, defaulting to
// ~/.ssh/tuikit_host_ed25519. The parent directory is created if absent.
func resolveHostKeyPath(p string) (string, error) {
	if p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home dir: %w", err)
	}
	dir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return filepath.Join(dir, "tuikit_host_ed25519"), nil
}
