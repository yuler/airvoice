class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.3.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.2/airvoice-cli-darwin-arm64"
      sha256 "5859a83a06cbe684284b13c076e68c19d132295da20c1815af6e755ebc86e187"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.2/airvoice-cli-darwin-amd64"
      sha256 "a4792637d5f8fc36265c33414a9d7ba145af49756fdb8ef9e21521dab2227818"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.2/airvoice-cli-linux-arm64"
      sha256 "a28c93adacbf0b471cf372dd6a4b6dc0608ba0cfb356eb559bc8b0ea0cfa6bc2"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.2/airvoice-cli-linux-amd64"
      sha256 "19297697c9ebd63ef9dde0a7f5d8cf21884b1932be68cedc0d5837cc953e8a2f"
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
