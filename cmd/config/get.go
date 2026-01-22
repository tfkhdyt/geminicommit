/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a configuration value.

[api]
  api.key             - Gemini API key
  api.model           - Gemini model name
  api.baseurl         - Custom base URL for Gemini API

[commit]
  commit.language     - Language for commit messages
  commit.max_length   - Maximum length of commit message

[behavior]
  behavior.stage_all   - Stage all changes in tracked files
  behavior.auto_select - Let AI select files and generate commit message
  behavior.no_confirm  - Skip confirmation prompt
  behavior.quiet       - Suppress output
  behavior.push        - Push committed changes to remote
  behavior.dry_run     - Run without making changes
  behavior.show_diff   - Show diff before committing
  behavior.no_verify   - Skip git commit-msg hook verification

Example:
  gmc config get commit.language
  gmc config get api.model`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		_, valid := ValidConfigKeys[key]
		if !valid {
			fmt.Printf("Error: unknown config key '%s'\n", key)
			fmt.Println("Run 'gmc config get --help' to see available keys")
			os.Exit(1)
		}

		value := viper.Get(key)
		if value == nil {
			fmt.Printf("%s = (not set)\n", key)
		} else {
			fmt.Printf("%s = %v\n", key, value)
		}
	},
}

func init() {
	ConfigCmd.AddCommand(getCmd)
}
