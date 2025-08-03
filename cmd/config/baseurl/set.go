/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package baseurl

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set {base_url}",
	Short: "Set custom base URL for Google Gemini API",
	Long:  `Set custom base URL for Google Gemini API`,
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl := args[0]
		viper.Set("api.baseurl", baseUrl)
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
