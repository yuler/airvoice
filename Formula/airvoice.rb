class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.2.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-darwin-arm64"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-darwin-amd64"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-linux-arm64"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-linux-amd64"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    if OS.mac?
      binary = Hardware::CPU.arm? ? "airvoice-cli-darwin-arm64" : "airvoice-cli-darwin-amd64"
    else
      binary = Hardware::CPU.arm? ? "airvoice-cli-linux-arm64" : "airvoice-cli-linux-amd64"
    end
    bin.install binary => "airvoice"
  end

  test do
    assert_match "airvoice", shell_output("#{bin}/airvoice version")
  end
end
