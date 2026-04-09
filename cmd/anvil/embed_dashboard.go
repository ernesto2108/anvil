//go:build dashboard

package main

import (
	anvilroot "github.com/ernesto2108/anvil"
	"github.com/ernesto2108/anvil/internal/cli"
)

func init() {
	// Wire the embedded frontend assets (declared in frontendfs.go at the module
	// root) into the CLI layer before cli.Run is called. The root-level package
	// is the only location where Go's embed can reference frontend/dist without
	// the ../ prefix that the embed directive forbids.
	cli.DashboardAssets = anvilroot.DashboardFS
}
