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

var ValidConfigKeys = map[string]bool{
	"api.key": true, "api.model": true, "api.baseurl": true,
	"commit.language": true, "commit.max_length": true,
	"behavior.stage_all": true, "behavior.auto_select": true,
	"behavior.no_confirm": true, "behavior.quiet": true,
	"behavior.push": true, "behavior.dry_run": true,
	"behavior.show_diff": true, "behavior.no_verify": true,
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

[api]
  api.key             - Gemini API key
  api.model           - Gemini model name (default: gemini-2.5-flash)
  api.baseurl         - Custom base URL for Gemini API

[commit]
  commit.language     - Language for commit messages (default: english)
  commit.max_length   - Maximum length of commit message (default: 72)

[behavior]
  behavior.stage_all   - Stage all changes in tracked files (default: false)
  behavior.auto_select - Let AI select files and generate commit message (default: false)
  behavior.no_confirm  - Skip confirmation prompt (default: false)
  behavior.quiet       - Suppress output (default: false)
  behavior.push        - Push committed changes to remote (default: false)
  behavior.dry_run     - Run without making changes (default: false)
  behavior.show_diff   - Show diff before committing (default: false)
  behavior.no_verify   - Skip git commit-msg hook verification (default: false)

Example:
  gmc config set commit.language korean
  gmc config set commit.max_length 100
  gmc config set behavior.push true`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		if !ValidConfigKeys[key] {
			fmt.Printf("Error: unknown config key '%s'\n", key)
			fmt.Println("Run 'gmc config set --help' to see available keys")
			os.Exit(1)
		}

		viper.Set(key, value)
		if err := viper.WriteConfig(); err != nil {
			fmt.Printf("Error: failed to write config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s = %v\n", key, value)
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}
