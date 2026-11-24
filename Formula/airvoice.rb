class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.4/airvoice-cli-darwin-arm64"
      sha256 "3f6e4282427b1f74e31caf00d981c9c2e804c0271805d29efb2c43381fe1978f"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.4/airvoice-cli-darwin-amd64"
      sha256 "19fd14371281b9c1472f1f8b71a3c82158759ac86030887141efcf1a6e135c2a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.4/airvoice-cli-linux-arm64"
      sha256 "9345b5e27750f763f5209f97b7d0868a84ef56dcb939b8f45f85fd8898626eb7"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.4/airvoice-cli-linux-amd64"
      sha256 "a1fe2e985ff831c16db844a67389832a62f6d28befd26089589cdfb3b8caffd8"
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
