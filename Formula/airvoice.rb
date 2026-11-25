class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.11"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.11/airvoice-cli-darwin-arm64"
      sha256 "178298520015bf25eebb5a3e137874f816d519bbb99ac6de071d9c6b463c5a14"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.11/airvoice-cli-darwin-amd64"
      sha256 "c1f39ca9d63059e1b9c58e078a48c027b1284d08f82243a2e68c3e6592c28022"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.11/airvoice-cli-linux-arm64"
      sha256 "5aa3fee3fa65e1459effab7d8e7cbccf33155e1d5c077ac1eb777eb360db1622"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.11/airvoice-cli-linux-amd64"
      sha256 "5167cb1d0c3e927c8ab2124c5d68c8527e838494665b8046f27ac16fc54ae5ba"
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
