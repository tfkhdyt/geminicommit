/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package key

import (
	"fmt"

	"github.com/spf13/cobra"
)

// KeyCmd represents the key command
var KeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage Google Gemini API key",
	Long:  `Manage Google Gemini API key`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("key called")
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
