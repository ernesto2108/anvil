//go:build dashboard

// Package anvil is the module root package.
// This file provides the embedded frontend assets for the Anvil Dashboard.
// It lives at the repo root so the embed path "frontend/dist" resolves correctly
// without ../ (which Go's embed directive does not permit).
package anvil

import "embed"

// DashboardFS holds the pre-built React frontend produced by `npm run build`.
// Pass this to dashboard.Run() via the CLI layer.
//
//go:embed all:frontend/dist
var DashboardFS embed.FS
