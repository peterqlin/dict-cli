package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/mweb/internal/output"
)

var antCmd = &cobra.Command{
	Use:   "ant <word>",
	Short: "Look up antonyms for a word",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.APIKeyThesaurus == "" {
			return fmt.Errorf("thesaurus API key not set\n\nProvide it via MWEB_API_KEY_THESAURUS or set api_key_thesaurus in ~/.config/mweb/config.yaml")
		}
		word := args[0]
		client := newAPIClient(cfg.APIKeyDict, cfg.APIKeyThesaurus)
		entries, err := client.Thesaurus(word)
		if err != nil {
			return err
		}
		return output.PrintAntonyms(cmd.Root().OutOrStdout(), word, entries, resolveFormat())
	},
}
