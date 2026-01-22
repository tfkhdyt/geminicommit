/*
Copyright © 2024 Taufik Hidayat <tfkhdyt@proton.me>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tfkhdyt/geminicommit/cmd/config"
	"github.com/tfkhdyt/geminicommit/internal/delivery/cli/handler"
	"github.com/tfkhdyt/geminicommit/internal/service"
)

var (
	cfgFile       string
	stageAll      = false
	autoSelect    = false
	userContext   string
	model         string
	noConfirm     = false
	quiet         = false
	push          = false
	dryRun        = false
	showDiff      = false
	maxLength     = 72
	language      = "english"
	issue         string
	noVerify      = false
	customBaseUrl string
	rootHandler   = handler.NewRootHandler()
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "gmc",
	Short:   "CLI that writes your git commit messages for you with Google Gemini AI",
	Long:    "CLI that writes your git commit messages for you with Google Gemini AI",
	Version: "0.6.1",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Apply config values to variables if flags are not explicitly set
		applyConfigDefaults(cmd)
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: rootHandler.RootCommand(
		context.Background(),
		&stageAll,
		&autoSelect,
		&userContext,
		&model,
		&noConfirm,
		&quiet,
		&push,
		&dryRun,
		&showDiff,
		&maxLength,
		&language,
		&issue,
		&noVerify,
		&customBaseUrl,
	),
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
		BoolVarP(&stageAll, "all", "a", stageAll, "stage all changes in tracked files")
	RootCmd.Flags().
		BoolVarP(&autoSelect, "auto", "", autoSelect, "let AI select files and generate commit message")
	RootCmd.Flags().
		BoolVarP(&noConfirm, "yes", "y", noConfirm, "skip confirmation prompt")
	RootCmd.Flags().
		BoolVarP(&quiet, "quiet", "q", quiet, "suppress output (only works with --yes)")
	RootCmd.Flags().
		BoolVarP(&push, "push", "p", push, "push committed changes to remote repository")
	RootCmd.Flags().
		StringVarP(&userContext, "context", "c", "", "additional context to be added to the commit message")
	RootCmd.Flags().
		StringVarP(&model, "model", "m", service.DefaultModel, "google gemini model to use")
	RootCmd.Flags().
		BoolVarP(&dryRun, "dry-run", "", dryRun, "run the command without making any changes")
	RootCmd.Flags().
		BoolVarP(&showDiff, "show-diff", "", showDiff, "show the diff before committing")
	RootCmd.Flags().
		IntVarP(&maxLength, "max-length", "l", maxLength, "maximum length of the commit message")
	RootCmd.Flags().
		StringVarP(&language, "language", "", language, "language of the commit message")
	RootCmd.Flags().
		StringVarP(&issue, "issue", "i", "", "issue number or title")
	RootCmd.Flags().
		BoolVarP(&noVerify, "no-verify", "", noVerify, "skip git commit-msg hook verification")
	RootCmd.Flags().
		StringVarP(&customBaseUrl, "baseurl", "", service.DefaultBaseUrl, "specify custom url for Google Gemini Pro API")

	// Bind flags to viper config keys
	// [api]
	viper.BindPFlag("api.model", RootCmd.Flags().Lookup("model"))
	viper.BindPFlag("api.baseurl", RootCmd.Flags().Lookup("baseurl"))
	// [commit]
	viper.BindPFlag("commit.language", RootCmd.Flags().Lookup("language"))
	viper.BindPFlag("commit.max_length", RootCmd.Flags().Lookup("max-length"))
	// [behavior]
	viper.BindPFlag("behavior.stage_all", RootCmd.Flags().Lookup("all"))
	viper.BindPFlag("behavior.auto_select", RootCmd.Flags().Lookup("auto"))
	viper.BindPFlag("behavior.no_confirm", RootCmd.Flags().Lookup("yes"))
	viper.BindPFlag("behavior.quiet", RootCmd.Flags().Lookup("quiet"))
	viper.BindPFlag("behavior.push", RootCmd.Flags().Lookup("push"))
	viper.BindPFlag("behavior.dry_run", RootCmd.Flags().Lookup("dry-run"))
	viper.BindPFlag("behavior.show_diff", RootCmd.Flags().Lookup("show-diff"))
	viper.BindPFlag("behavior.no_verify", RootCmd.Flags().Lookup("no-verify"))
}

// applyConfigDefaults applies config values to variables if flags are not explicitly set
func applyConfigDefaults(cmd *cobra.Command) {
	flags := cmd.Flags()

	// [api]
	if !flags.Changed("model") && viper.IsSet("api.model") {
		model = viper.GetString("api.model")
	}
	if !flags.Changed("baseurl") && viper.IsSet("api.baseurl") {
		customBaseUrl = viper.GetString("api.baseurl")
	}
	// [commit]
	if !flags.Changed("language") && viper.IsSet("commit.language") {
		language = viper.GetString("commit.language")
	}
	if !flags.Changed("max-length") && viper.IsSet("commit.max_length") {
		maxLength = viper.GetInt("commit.max_length")
	}
	// [behavior]
	if !flags.Changed("all") && viper.IsSet("behavior.stage_all") {
		stageAll = viper.GetBool("behavior.stage_all")
	}
	if !flags.Changed("auto") && viper.IsSet("behavior.auto_select") {
		autoSelect = viper.GetBool("behavior.auto_select")
	}
	if !flags.Changed("yes") && viper.IsSet("behavior.no_confirm") {
		noConfirm = viper.GetBool("behavior.no_confirm")
	}
	if !flags.Changed("quiet") && viper.IsSet("behavior.quiet") {
		quiet = viper.GetBool("behavior.quiet")
	}
	if !flags.Changed("push") && viper.IsSet("behavior.push") {
		push = viper.GetBool("behavior.push")
	}
	if !flags.Changed("dry-run") && viper.IsSet("behavior.dry_run") {
		dryRun = viper.GetBool("behavior.dry_run")
	}
	if !flags.Changed("show-diff") && viper.IsSet("behavior.show_diff") {
		showDiff = viper.GetBool("behavior.show_diff")
	}
	if !flags.Changed("no-verify") && viper.IsSet("behavior.no_verify") {
		noVerify = viper.GetBool("behavior.no_verify")
	}
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
