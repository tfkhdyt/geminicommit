/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package baseurl

import (
	"github.com/spf13/cobra"
)

// BaseurlCmd represents the baseurl command
var BaseurlCmd = &cobra.Command{
	Use:   "baseurl",
	Short: "Manage custom base URL for Google Gemini API",
	Long:  `Manage custom base URL for Google Gemini API`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("baseurl called")
	},
}

func init() {
	BaseurlCmd.AddCommand(setCmd, showCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// baseurlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// baseurlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
