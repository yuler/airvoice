class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.2.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.2.3/airvoice-cli-darwin-arm64"
      sha256 "4aa61f049dd38ed348a7d3fe0a8abeb682ca36e1931ae3d2ec3e071d839f5cdd"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.2.3/airvoice-cli-darwin-amd64"
      sha256 "ed8e5b314c5b9918cf2e4049319f3fb4a6b52aaa4360a5bd43dee03678b02ebc"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.2.3/airvoice-cli-linux-arm64"
      sha256 "d7cb9d599eee1adb91733d1b216ba52b4f025d536f38f4637fdb52d475817d57"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.2.3/airvoice-cli-linux-amd64"
      sha256 "4558e29cf792803bc3d04c4f3cda1839b6ecbd9ae93ddc57f44210b7e4c2ae27"
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
