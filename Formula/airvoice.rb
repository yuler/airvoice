class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.2.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-darwin-arm64"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-darwin-amd64"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-linux-arm64"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-linux-amd64"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    if OS.mac?
      binary = Hardware::CPU.arm? ? "airvoice-darwin-arm64" : "airvoice-darwin-amd64"
    else
      binary = Hardware::CPU.arm? ? "airvoice-linux-arm64" : "airvoice-linux-amd64"
    end
    bin.install binary => "airvoice"
  end

  test do
    assert_match "airvoice", shell_output("#{bin}/airvoice version")
  end
end
