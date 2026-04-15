//go:build !dashboard

package cli

import (
	"os"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdDashboard(_ *config.App, _ []string) {
	output.Error("dashboard not available in this build")
	output.Info("rebuild with dashboard support: wails build")
	os.Exit(1)
}
