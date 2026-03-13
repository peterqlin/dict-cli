package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/mweb/internal/output"
)

var defCmd = &cobra.Command{
	Use:   "def <word>",
	Short: "Look up the definition of a word",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.APIKeyDict == "" {
			return fmt.Errorf("dictionary API key not set\n\nProvide it via MWEB_API_KEY_DICT or set api_key_dict in ~/.config/mweb/config.yaml")
		}
		word := args[0]
		client := newAPIClient(cfg.APIKeyDict, cfg.APIKeyThesaurus)
		entries, err := client.Define(word)
		if err != nil {
			return err
		}
		return output.PrintDefinitions(cmd.Root().OutOrStdout(), word, entries, cfg.MaxDefinitions, resolveFormat())
	},
}
