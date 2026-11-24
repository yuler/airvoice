class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.6"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.6/airvoice-cli-darwin-arm64"
      sha256 "d81e23d902baf57ec19b2d56f1595d10202a8712ad6205c3b38ac4d0b8ea0bb9"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.6/airvoice-cli-darwin-amd64"
      sha256 "b4f9a0a18cf3785ebe2be01c3c9536261010548615c0c25c7d993aa91b958c0f"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.6/airvoice-cli-linux-arm64"
      sha256 "5715d4e0526bf6e144bbadd740be3b32a3ede50986317c513a6f5bfb3827bd02"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.6/airvoice-cli-linux-amd64"
      sha256 "680d26ea442272eabd754c2d3b9d6ee9086bd2a2a01c1a410070f456d4945d8c"
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
