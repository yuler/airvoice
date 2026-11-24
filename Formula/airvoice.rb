class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.8"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.8/airvoice-cli-darwin-arm64"
      sha256 "875be7be6795ed999fe89a803f3c96929d98060adda393f1d15141c425aec144"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.8/airvoice-cli-darwin-amd64"
      sha256 "7c963acd79401e8a56344effccad83d34cd28d1e422aab5583885562fbe0c2ab"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.8/airvoice-cli-linux-arm64"
      sha256 "7d6a002d0f97bf829d8acfb46886a12491e8d62508261350d4f7625f069204c6"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.8/airvoice-cli-linux-amd64"
      sha256 "c1c46a1e09dcddbb72a956bc7722074f1ae7db4492112df608accbf9d5ce862e"
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
