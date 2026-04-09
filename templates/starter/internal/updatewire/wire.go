// Package updatewire wires tuikit-go's auto-update system for myapp.
//
// Replace OWNER, REPO, and BINARY with your own values, then pass
// Config() into tuikit.WithAutoUpdate when building the app.
package updatewire

import (
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

// version is set at build time via ldflags:
//
//	-X github.com/OWNER/myapp/internal/updatewire.version=v1.2.3
var version = "dev"

// Config returns a tuikit.UpdateConfig for myapp.
// Version is injected at link time; in dev builds it stays "dev" and
// the update check is skipped automatically by tuikit-go.
func Config() tuikit.UpdateConfig {
	return tuikit.UpdateConfig{
		Owner:      "OWNER",
		Repo:       "myapp",
		BinaryName: "myapp",
		Version:    version,
		Mode:       tuikit.UpdateNotify,
	}
}
