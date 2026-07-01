class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.2.0/airvoice-darwin-arm64"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.2.0/airvoice-darwin-amd64"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.2.0/airvoice-linux-arm64"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.2.0/airvoice-linux-amd64"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    bin.install "airvoice"
  end

  test do
    assert_match "airvoice", shell_output("#{bin}/airvoice version")
  end
end
