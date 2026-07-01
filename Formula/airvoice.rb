class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.2.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-darwin-arm64"
      sha256 "a1bee32be54d92285ca0040b169fec8a6856ba7fbd3e28431589f65f1e3d5461"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-darwin-amd64"
      sha256 "d579081451456ac6c11d08f93801fd41e44d90f33f3887deca395283f26d5ace"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-linux-arm64"
      sha256 "3a46e5b61d1a40f551ab5511cffe34a9f8c0cfb6b86096bdd2ebb91613d52f57"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-cli-linux-amd64"
      sha256 "2742974c7c599e86be6affc9be19867fc4f60dbdc95257149534618bdcf7ed31"
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
