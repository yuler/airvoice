class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.7"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.7/airvoice-cli-darwin-arm64"
      sha256 "328cdd46d900d8af5ec7c1b5ff2ebd797cfd7e95362294d539fc321b4c56bf61"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.7/airvoice-cli-darwin-amd64"
      sha256 "f9ebf87087bd39accb23ad01b1bbb3d0f9820b04d8278d215e482611d60e7c71"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.7/airvoice-cli-linux-arm64"
      sha256 "d0b585cb8736c7fbd7bd73225e47a7f05ffbf4d1ce62406d939701651ed3c999"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.7/airvoice-cli-linux-amd64"
      sha256 "50f787b59498adeba2eb930e795ab8a86c036b566315bb4965e2d00fa6f53859"
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
