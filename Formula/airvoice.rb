class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.1/airvoice-cli-darwin-arm64"
      sha256 "ee01201fb3e8f767418ce70166d384c99df8a4773e20069c59df99ee1ddf500d"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.1/airvoice-cli-darwin-amd64"
      sha256 "b965901b8ef962edc1c84da2060deae04e33285a5921806286332fa67ec5e146"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.1/airvoice-cli-linux-arm64"
      sha256 "c1e46f6501a0367188d3635adb25e90c66aa06fbc78dcb473050cebbcc5b6658"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.1/airvoice-cli-linux-amd64"
      sha256 "4079f43ab1dd7e299cb8bd576ac180e89682fb195283d27169b43faf4bd83838"
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
