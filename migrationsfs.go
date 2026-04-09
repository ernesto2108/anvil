// Package anvil is the module root.
// This file embeds the database migrations so the binary is self-contained.
package anvil

import "embed"

// MigrationsFS holds the SQL migrations for the dashboard store.
// They are embedded so the binary does not depend on migrations existing
// on disk at runtime.
//
//go:embed all:migrations
var MigrationsFS embed.FS
