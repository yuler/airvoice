class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.5.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.5.0/airvoice-cli-darwin-arm64"
      sha256 "34e9f2b2f1c4b0e1f4ef9ba968c223eb46d6469eb564db06a410228fad0d006d"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.5.0/airvoice-cli-darwin-amd64"
      sha256 "f3e76bb8e619190b87286c742b5c1ece6a616ee1caea9b77cf88e8db247edc54"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.5.0/airvoice-cli-linux-arm64"
      sha256 "cc96c3c05b4ae62bd6802153c6d55722b3f208198d12b53fde647302c88dc1ea"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.5.0/airvoice-cli-linux-amd64"
      sha256 "0c476eba56c3c476ab064c55caa6becd16b2ffed5d453f008c8eb29ce9415e8c"
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
