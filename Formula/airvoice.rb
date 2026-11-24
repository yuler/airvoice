class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.2/airvoice-cli-darwin-arm64"
      sha256 "d56ac1ca38e62200632445b00b33089926a5d37a7bef06d6279da74699704e57"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.2/airvoice-cli-darwin-amd64"
      sha256 "0201705931f9481247ce9aa7345ce8e8bf5fb6e280e9fdc589402298f13389a1"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.2/airvoice-cli-linux-arm64"
      sha256 "dacb0ae4aee5e3182e65c44977c1cfbf98a5caded00949236bdf70a6ea59eede"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.2/airvoice-cli-linux-amd64"
      sha256 "72d790b8bbfc687f7a0eb9bc0502f18aeaa8f31d720dcb881a448e2df4f203eb"
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
