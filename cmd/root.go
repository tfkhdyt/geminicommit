/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tfkhdyt/geminicommit/cmd/config"
	"github.com/tfkhdyt/geminicommit/internal/gemini"
	"github.com/tfkhdyt/geminicommit/pkg/git"
)

var (
	cfgFile  string
	stageAll bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "geminicommit",
	Short:   "A CLI that writes your git commit messages for you with Google Gemini AI ",
	Long:    `A CLI that writes your git commit messages for you with Google Gemini AI `,
	Version: "0.0.1",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(_ *cobra.Command, _ []string) {
		cobra.CheckErr(git.VerifyGitInstallation())
		cobra.CheckErr(git.VerifyGitRepository())

		if stageAll {
			cobra.CheckErr(git.StageAll())
		}

		filesChan := make(chan []string, 1)
		diffChan := make(chan string, 1)

		cobra.CheckErr(
			spinner.New().
				Title("Detecting staged files").
				Action(git.DetectDiffChanges(filesChan, diffChan)).
				Run(),
		)
		files, diff := <-filesChan, <-diffChan

		if len(files) == 0 {
			fmt.Println(
				"Error: No staged changes found. Stage your changes manually, or automatically stage all changes with the `--all` flag",
			)
			os.Exit(1)
		} else if len(files) == 1 {
			fmt.Printf("Detected %d staged file:\n", len(files))
		} else {
			fmt.Printf("Detected %d staged files:\n", len(files))
		}

		for idx, file := range files {
			fmt.Printf("     %d. %s\n", idx+1, file)
		}

		messageChan := make(chan string, 1)
		cobra.CheckErr(
			spinner.New().
				Title("The AI is analyzing your changes").
				Action(gemini.AnalyzeChanges(diff, messageChan)).
				Run(),
		)

		message := <-messageChan
		fmt.Println("\nChanges analyzed!")

		if strings.TrimSpace(message) == "" {
			fmt.Println("Error: no commit messages were generated. Try again")
			os.Exit(1)
		}

		fmt.Printf("\n%s\n\n", message)

		var confirm bool
		cobra.CheckErr(huh.NewConfirm().
			Title("Use this commit message?").
			Affirmative("Yes!").
			Negative("No.").
			Value(&confirm).Run())

		if !confirm {
			fmt.Println("Commit cancelled")
			return
		}

		cobra.CheckErr(git.CommitChanges(message))

		fmt.Println("✔ Successfully committed!")
	},
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
		viper.SetDefault("api.key", "")

		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			cobra.CheckErr(viper.WriteConfig())
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error: failed to read config")
		os.Exit(1)
	}
}
