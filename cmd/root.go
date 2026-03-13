package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/mweb/internal/api"
	"github.com/user/mweb/internal/config"
	"github.com/user/mweb/internal/output"
)

var (
	cfg        *config.Config
	jsonOutput bool
)

// newAPIClient is the factory used to create API clients. Overrideable in tests.
var newAPIClient = api.NewClient

var rootCmd = &cobra.Command{
	Use:   "mweb",
	Short: "Query the Merriam-Webster dictionary and thesaurus",
	Long: `mweb is a CLI for the Merriam-Webster Collegiate Dictionary and Thesaurus.

API keys are read from ~/.config/mweb/config.yaml or from environment variables:
  MWEB_API_KEY_DICT      — Collegiate Dictionary key
  MWEB_API_KEY_THESAURUS — Collegiate Thesaurus key`,
	// Silence default error printing; we handle it in Execute().
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	// Load config before any subcommand runs.
	// Skip if cfg is already set (e.g., injected by tests).
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cfg != nil {
			return nil
		}
		var err error
		cfg, err = config.Load()
		return err
	}

	rootCmd.AddCommand(defCmd, synCmd, antCmd)
}

// resolveFormat returns the output format, with --json flag taking precedence.
func resolveFormat() output.Format {
	if jsonOutput {
		return output.FormatJSON
	}
	if cfg != nil && cfg.OutputFormat == "json" {
		return output.FormatJSON
	}
	return output.FormatPlain
}
