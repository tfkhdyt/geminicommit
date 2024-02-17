/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"

	"github.com/tfkhdyt/geminicommit/cmd/config"
)

var (
	cfgFile  string
	stageAll bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "geminicommit",
	Short: "A CLI that writes your git commit messages for you with Google Gemini AI ",
	Long:  `A CLI that writes your git commit messages for you with Google Gemini AI `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := exec.Command("git", "-v").Output(); err != nil {
			fmt.Println("Git is not installed")
			os.Exit(1)
		}

		if _, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err != nil {
			fmt.Println("The current directory must be a Git repository")
			os.Exit(1)
		}

		if stageAll {
			if _, err := exec.Command("git", "add", "-u").Output(); err != nil {
				fmt.Println("Failed to update tracked files")
				os.Exit(1)
			}
		}

		filesChan := make(chan []string, 1)
		diffChan := make(chan string, 1)

		if err := spinner.New().Title("Detecting staged files").Action(func() {
			files, err := exec.Command("git", "diff", "--cached", "--diff-algorithm=minimal", "--name-only").Output()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			filesStr := strings.TrimSpace(string(files))

			if filesStr == "" {
				filesChan <- []string{}
				diffChan <- ""
				return
			}

			diff, err := exec.Command("git", "diff", "--cached", "--diff-algorithm=minimal").Output()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			filesChan <- strings.Split(filesStr, "\n")
			diffChan <- string(diff)
		}).Run(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		files, diff := <-filesChan, <-diffChan

		if len(files) == 0 {
			fmt.Println("No staged changes found. Stage your changes manually, or automatically stage all changes with the `--all` flag")
			os.Exit(1)
		} else if len(files) == 1 {
			fmt.Printf("Detected %d staged file:\n", len(files))
		} else {
			fmt.Printf("Detected %d staged files:\n", len(files))
		}

		for _, file := range files {
			fmt.Println("    ", file)
		}

		messageChan := make(chan string, 1)
		if err := spinner.New().Title("The AI is analyzing your changes").Action(func() {
			ctx := context.Background()

			client, err := genai.NewClient(ctx, option.WithAPIKey(viper.GetString("api.key")))
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			defer client.Close()

			model := client.GenerativeModel("gemini-pro")
			resp, err := model.GenerateContent(ctx, genai.Text(fmt.Sprintf(`generate commit messages based on the following diff:
%s

commit messages should:
 - follow conventional commits
 - message format should be: <type>[scope]: <description>

examples:
 - fix(authentication): add password regex pattern
 - feat(storage): add new test cases`, diff)))
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			messageChan <- fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
		}).Run(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		message := <-messageChan
		fmt.Println("\nChanges analyzed!")

		if strings.TrimSpace(message) == "" {
			fmt.Println("No commit messages were generated. Try again")
			os.Exit(1)
		}

		fmt.Printf("\n%s\n\n", message)

		var confirm bool
		if err := huh.NewConfirm().
			Title("Use this commit message?").
			Affirmative("Yes!").
			Negative("No.").
			Value(&confirm).Run(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if !confirm {
			fmt.Println("Commit cancelled")
			return
		}

		if err := exec.Command("git", "commit", "-m", message).Run(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/geminicommit/config.toml)")
	RootCmd.Flags().BoolVarP(&stageAll, "all", "a", false, "stage all changes in tracked files (default is false)")
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
