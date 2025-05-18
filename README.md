[![Support Palestine](https://raw.githubusercontent.com/Safouene1/support-palestine-banner/master/banner-project.svg)](https://github.com/Safouene1/support-palestine-banner)

# geminicommit

**AI-Powered, Conventional Commit Messages with Google Gemini**

![Preview](./assets/Screenshot_20241112_103154.png)

**geminicommit** helps you write clear, conventional, and meaningful Git commit messages automatically using Google Gemini AI. Save time, improve your commit history, and focus on what matters—your code.

---

## ✨ Features

- **AI-Generated Commit Messages:** Let Gemini AI analyze your staged changes and suggest concise, descriptive commit messages.
- **Customizable Output:** Tailor the message style and structure to fit your workflow.
- **Conventional Commits:** Ensures messages follow best practices for readability and automation.
- **Cross-Platform:** Works on Linux, Windows, and macOS.
- **Open Source:** Free to use and contribute.

---

## 🚀 Quickstart

```sh
# 1. Install (Go required)
go install github.com/tfkhdyt/geminicommit@latest

# 2. Get your Gemini API key
#    https://aistudio.google.com/app/apikey

# 3. Configure your API key
geminicommit config key set <your-api-key>

# 4. Stage your changes
git add <file>

# 5. Generate and commit
geminicommit
```

---

## 🛠️ Installation

- **From Source:**
  ```sh
  go install github.com/tfkhdyt/geminicommit@latest
  ```
- **Standalone Binary:**
  Download from the [releases page](https://github.com/tfkhdyt/geminicommit/releases) and move to a directory in your `PATH`:
  - Linux: `$HOME/.local/bin/` or `/usr/local/bin/`
  - Windows: `%LocalAppData%\Programs\`
  - macOS: `/usr/local/bin/`
- **NixOS:**
  ```nix
  environment.systemPackages = [
    pkgs.geminicommit
  ];
  ```

---

## ⚙️ Configuration

1. Get your API key from [Google AI Studio](https://aistudio.google.com/app/apikey).
2. Set your key:
   ```sh
   geminicommit config key set <your-api-key>
   ```

---

## 📖 Usage

1. Stage your changes:
   ```sh
   git add <file>
   ```
2. Run geminicommit:
   ```sh
   geminicommit
   ```
3. Review and edit the AI-generated message if needed.
4. geminicommit will commit your changes with the generated message.

For more options:

```sh
geminicommit --help
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to open an issue or submit a pull request.

---

## 📄 License

This project is licensed under the GPLv3 License. See the [LICENSE](LICENSE) file for details.
