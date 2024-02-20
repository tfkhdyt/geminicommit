[![ReadMeSupportPalestine](https://raw.githubusercontent.com/Safouene1/support-palestine-banner/master/banner-project.svg)](https://github.com/Safouene1/support-palestine-banner)

# geminicommit: Write Clear, Concise Git Commit Messages with Google Gemini AI

**Tired of writer's block crafting commit messages?** `geminicommit` harnesses the
power of Google's Gemini AI to help you write **informative, concise, and
well-formatted commit messages** effortlessly, streamlining your Git workflow.

## Key Features

- **AI-powered Message Generation:** Leverage Google Gemini AI to generate clear
  and descriptive messages, saving you time and brainpower.
- **Customizable:** Tailor the message generation process to your specific needs
  and preferences.
- **Conventional Commits Compliant:** Adhere to widely accepted commit message
  formatting standards for better project readability and maintainability.
- **Cross-Platform Compatibility:** Works seamlessly on Linux, Windows, and macOS.
  systems.
- **Free and Open Source:** Contribute to and benefit from the open-source community.

## Getting Started

### Installation

- **Build from source:** `go install github.com/tfkhdyt/geminicommit`
- **Arch Linux (AUR):** `yay -S geminicommit-bin`
- **Standalone binary:** Download the binary file from
  [release page](https://github.com/tfkhdyt/geminicommit/releases) and move the
  binary to one of the `PATH` directories in your system, for example:
  - **Linux:** `$HOME/.local/bin/`, `/usr/local/bin/`
  - **Windows:** `%LocalAppData%\Programs\`
  - **macOS:** `/usr/local/bin/`

### Usage

- Stage your changes in Git `git add file_name.go`.
- Run `geminicommit` in your terminal.
- Review the AI-generated message and customize it as needed.
- `geminicommit` will automatically commit your changes with the generated
  message.

More details in `geminicommit --help`

## License

This project is licensed under the GPLv3 License. See the LICENSE file for details.
