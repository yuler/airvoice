class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.3.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.1/airvoice-cli-darwin-arm64"
      sha256 "c82eefcaa3b27da36139e5f708f650bcae20ffdb3d6a39f5d49a2a93aceeef1c"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.1/airvoice-cli-darwin-amd64"
      sha256 "8fc5657e5e7a84edbf0ed02ce50c8c5cc715fd71347fde983f190cd55bd03053"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.1/airvoice-cli-linux-arm64"
      sha256 "5e8b2225f9c816304ca35fb0e16754aa9fd620c85ddcbecf772bd43b6eee19e2"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.1/airvoice-cli-linux-amd64"
      sha256 "42f52995f74a3c6fdb214f8b98a0c9198433da4bbe29913c36fe59d842e8fb4d"
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
