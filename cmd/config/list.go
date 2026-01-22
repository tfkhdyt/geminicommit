/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Long:  `List all configuration values`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()
		printSettings(settings, "")
	},
}

func printSettings(settings map[string]interface{}, prefix string) {
	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case map[string]interface{}:
			printSettings(v, fullKey)
		default:
			fmt.Printf("%s = %v\n", fullKey, v)
		}
	}
}

func init() {
	ConfigCmd.AddCommand(listCmd)
}
