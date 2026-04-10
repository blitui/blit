// Package updatewire wires blit's auto-update system for myapp.
//
// Replace OWNER, REPO, and BINARY with your own values, then pass
// Config() into blit.WithAutoUpdate when building the app.
package updatewire

import (
	blit "github.com/blitui/blit"
)

// version is set at build time via ldflags:
//
//	-X github.com/OWNER/myapp/internal/updatewire.version=v1.2.3
var version = "dev"

// Config returns a blit.UpdateConfig for myapp.
// Version is injected at link time; in dev builds it stays "dev" and
// the update check is skipped automatically by blit.
func Config() blit.UpdateConfig {
	return blit.UpdateConfig{
		Owner:      "OWNER",
		Repo:       "myapp",
		BinaryName: "myapp",
		Version:    version,
		Mode:       blit.UpdateNotify,
	}
}
