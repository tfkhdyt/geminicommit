/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package baseurl

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show currently configured custom base URL for Google Gemini API",
	Long:  `Show currently configured custom base URL for Google Gemini API`,
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := viper.GetString("api.baseurl")
		if baseUrl == "" {
			fmt.Println("No custom base URL configured")
		} else {
			fmt.Println(baseUrl)
		}
	},
}

func init() {
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
