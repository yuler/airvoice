class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.5.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.5.2/airvoice-cli-darwin-arm64"
      sha256 "ca97e18b4452acc82ed75427b58854530efc9509f7ac3dd1c0ba6bb2f324fce7"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.5.2/airvoice-cli-darwin-amd64"
      sha256 "dfc5fe3b4f4e2471178e0f4e9ed659cfbb3f9bebbfb4ed70463dc242466e71f3"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.5.2/airvoice-cli-linux-arm64"
      sha256 "9c2df2f5c1628b96174410717d58c049bb8eedc32fb29d8f25882a8f6a60ff2a"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.5.2/airvoice-cli-linux-amd64"
      sha256 "17c0d7c26b7d07ea7f6dcb2e859f937a46b7d02513ff1e0e90a6aadcef9b14d2"
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
