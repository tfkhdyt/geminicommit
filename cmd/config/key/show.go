/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package key

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// showCmd represents the set command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show currently used Google Gemini API key",
	Long:  `Show currently used Google Gemini API key`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := viper.GetString("api.key")
		fmt.Println(apiKey)
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
