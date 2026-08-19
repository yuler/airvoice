class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.0/airvoice-cli-darwin-arm64"
      sha256 "d7d3ae1f63852b35ad0a6626f576aaf4f6a351cbb7f2e02e305eb0d839888f24"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.0/airvoice-cli-darwin-amd64"
      sha256 "c4fdc48fd6e632ff917d86252a9561de718bb66358a9ae002ae1f976da4d8b49"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.0/airvoice-cli-linux-arm64"
      sha256 "858f5a1f48168480e849214f97d9ab66fe542f33c0e513f274e39211c33f1a4f"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.0/airvoice-cli-linux-amd64"
      sha256 "c4b2311d88aed9cfa595df4c976f7fca7cc59d4b529ee0b68a23709f237fec13"
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
