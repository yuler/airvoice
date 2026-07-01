class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.2.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-darwin-arm64"
      sha256 "b08cfa5a09346dfc62f7f9d009892345f3a1137c762a494badd254282308d054"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-darwin-amd64"
      sha256 "1ffcd3c7fb30bdfa12ffb562ecc2eeaa9ddf48b9dd690196f697361db3b65925"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-linux-arm64"
      sha256 "da1c8185abb69ca4cfded32cb6596f97080c2809ed843506de6076a3fa046be1"
    else
      url "https://github.com/yuler/airvoice/releases/download/v#{version}/airvoice-linux-amd64"
      sha256 "6c2cc917f4c7de8f3916c438846885d3c001a4051a212ba19e56277e289610b8"
    end
  end

  def install
    if OS.mac?
      binary = Hardware::CPU.arm? ? "airvoice-darwin-arm64" : "airvoice-darwin-amd64"
    else
      binary = Hardware::CPU.arm? ? "airvoice-linux-arm64" : "airvoice-linux-amd64"
    end
    bin.install binary => "airvoice"
  end

  test do
    assert_match "airvoice", shell_output("#{bin}/airvoice version")
  end
end
