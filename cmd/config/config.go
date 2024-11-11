/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package config

import (
	"github.com/spf13/cobra"

	"github.com/tfkhdyt/geminicommit/cmd/config/key"
	"github.com/tfkhdyt/geminicommit/cmd/config/model"
)

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage geminicommit configuration through cli",
	Long:  `Manage geminicommit configuration through cli`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("config called")
	},
}

func init() {
	ConfigCmd.AddCommand(key.KeyCmd, model.KeyCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
