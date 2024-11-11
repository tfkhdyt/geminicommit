/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package model

import (
	"github.com/spf13/cobra"
)

// KeyCmd represents the key command
var KeyCmd = &cobra.Command{
	Use:   "model",
	Short: "Choose Google Gemini Model",
	Long:  `Choose Google Gemini Model (Docs: https://ai.google.dev/gemini-api/docs/models/gemini)`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("key called")
	},
}

func init() {
	KeyCmd.AddCommand(setCmd, showCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// keyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// keyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
