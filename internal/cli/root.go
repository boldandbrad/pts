package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var version = "0.0.1-alpha.2"

var rootCmd = &cobra.Command{
	Use:     "pts",
	Version: version,
	Short:   "Pick the Stick points calculator",
	Long: `pts is a Pick the Stick points calculator for Major League Baseball.
	It allows you to easily compare players to make the best pick in your draft.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
