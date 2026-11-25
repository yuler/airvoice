class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.10"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.10/airvoice-cli-darwin-arm64"
      sha256 "7259e959059b7c60ba10bbd73d24afd603fe43be5a8cfb6226b2149dffad0b95"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.10/airvoice-cli-darwin-amd64"
      sha256 "3ea3eb503ac799357349668b6f1fdb41c5166b79d58ba01ce7156f1190f1138f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.10/airvoice-cli-linux-arm64"
      sha256 "e2bfb910597a111bfcdab30679ddc0c3ba415048a006bc236f52e32e08e58bb6"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.10/airvoice-cli-linux-amd64"
      sha256 "bdfde2ec0daf42098ea24a5a1f0bf2157040caab2779ec3020dd8029548b443e"
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
