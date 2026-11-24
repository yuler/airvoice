class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.5"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.5/airvoice-cli-darwin-arm64"
      sha256 "778ec824f7af284e56fc2247743ab083d00ccabfce46416b3b78c3c750181d6f"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.5/airvoice-cli-darwin-amd64"
      sha256 "193f70a103e7f63a244ceaa4c6037d28e01d01fd61dbe8f044afa054597992e1"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.5/airvoice-cli-linux-arm64"
      sha256 "87fe1744675461936a15ff4a6341027591da07212174189fe02e673d847e5538"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.5/airvoice-cli-linux-amd64"
      sha256 "d06fc8fba991af8218c00e5930c542ba9977e39e432951b28be3a7f6cc730769"
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
