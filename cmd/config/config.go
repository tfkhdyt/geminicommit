/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package config

import (
	"github.com/spf13/cobra"
)

// ConfigCmd represents the config command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage geminicommit configuration through cli",
	Long:  `Manage geminicommit configuration through cli`,
}
