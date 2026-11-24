class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.3/airvoice-cli-darwin-arm64"
      sha256 "92981b49dcb1fbc2bf2d78343eb25acbd9f31fd7ae35b2bc2ce598463815e5fa"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.3/airvoice-cli-darwin-amd64"
      sha256 "ac4735bbd2aebd7b15574035eded6416e773c852b7574c4c7c9f9e47080dec91"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.3/airvoice-cli-linux-arm64"
      sha256 "082ea50cde82e4b4e0b23edf6c94f10efde93a6b659218d46b85fac46a3c2277"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.3/airvoice-cli-linux-amd64"
      sha256 "67e36a55a5d8ef20999ebb9684d3d8bff2d47b57d60aba5b4f36c1fe12374963"
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
