class Geminicommit < Formula
  desc "A CLI that writes your git commit messages for you with Google Gemini AI"
  homepage "https://github.com/tfkhdyt/geminicommit"
  version "0.0.3"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/tfkhdyt/geminicommit/releases/download/v0.0.3/geminicommit-v0.0.3-darwin-amd64.tar.gz"
      sha256 "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    else
      url "https://github.com/tfkhdyt/geminicommit/releases/download/v0.0.3/geminicommit-v0.0.3-darwin-arm64.tar.gz"
      sha256 "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    end
  end

  def install
    bin.install "geminicommit"
  end

  test do
    system "#{bin}/geminicommit", "--version"
  end
end
