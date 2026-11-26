class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.5.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.5.1/airvoice-cli-darwin-arm64"
      sha256 "57870266e3c79a71b6d7d2f600b9780cc64c9776d92b894ef9fc9eee26ddc932"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.5.1/airvoice-cli-darwin-amd64"
      sha256 "3a86fc82a6ed9243e3e399e136f693ab43023b04aa59095a767b1bdf964d4c42"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.5.1/airvoice-cli-linux-arm64"
      sha256 "609d54cc67300f0b45da675b7bd9ec44f8b682af78a21b55c8efab49ca72b0c5"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.5.1/airvoice-cli-linux-amd64"
      sha256 "551cab64da4bfb1748c7569c783c617760354158ec32f6a27c0281368b51f66f"
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
