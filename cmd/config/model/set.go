/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package model

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set {model_name}",
	Short: "Set Google Gemini model",
	Long:  `Set Google Gemini model`,
	Run: func(cmd *cobra.Command, args []string) {
		model := args[0]
		viper.Set("api.model", model)
		cobra.CheckErr(viper.WriteConfig())
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
