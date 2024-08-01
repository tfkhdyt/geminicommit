/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tfkhdyt/geminicommit/cmd/config"
	"github.com/tfkhdyt/geminicommit/internal/container"
)

var (
	cfgFile     string
	stageAll    bool
	userContext string
	rootHandler = container.GetRootHandlerInstance()
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "geminicommit",
	Short:   "A CLI that writes your git commit messages for you with Google Gemini AI ",
	Long:    `A CLI that writes your git commit messages for you with Google Gemini AI `,
	Version: "0.1.0",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: rootHandler.RootCommand(&stageAll, &userContext),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.AddCommand(config.ConfigCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/geminicommit/config.toml)")
	RootCmd.Flags().
		BoolVarP(&stageAll, "all", "a", false, "stage all changes in tracked files (default is false)")
	RootCmd.Flags().
		StringVarP(&userContext, "context", "c", "none", "additional context to be added to the commit message")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directory.
		config, err := os.UserConfigDir()
		cobra.CheckErr(err)
		configDirPath := filepath.Join(config, "geminicommit")
		configFilePath := filepath.Join(configDirPath, "config.toml")

		viper.AddConfigPath(configDirPath)
		viper.SetConfigType("toml")
		viper.SetConfigName("config")

		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			createConfig()
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error: failed to read config")
		os.Exit(1)
	}
}

func createConfig() {
	// Create the directory and file paths.
	config, err := os.UserConfigDir()
	cobra.CheckErr(err)
	configDirPath := filepath.Join(config, "geminicommit")
	configFilePath := filepath.Join(configDirPath, "config.toml")

	// Create the directory if it does not exist.
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDirPath, 0o755); err != nil {
			fmt.Println("Error: failed to make config dir")
			os.Exit(1)
		}
	}

	// Create the file if it does not exist.
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		file, err := os.Create(configFilePath)
		if err != nil {
			fmt.Println("Error: failed to make config file")
			os.Exit(1)
		}
		defer file.Close()
	}
}
