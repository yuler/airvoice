class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.3.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.0/airvoice-cli-darwin-arm64"
      sha256 "cbcd6071de4ce2d1f28ad60cff644963c64d777a19ba07be1efe8536df87010e"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.0/airvoice-cli-darwin-amd64"
      sha256 "458f13b59e67ae3afa8203ada4534109e8cbd10038d48dda264a58b054941abc"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.0/airvoice-cli-linux-arm64"
      sha256 "cbc19170ca341b4bda939451edfc0c6c00efec4e8a8ddafbda4c8a690540ba22"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.0/airvoice-cli-linux-amd64"
      sha256 "44c5ce03392a11c361be974aa807d07d1f12f0dd091e1da26525557ceb1e85ce"
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
