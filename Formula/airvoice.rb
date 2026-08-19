class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.3.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.3/airvoice-cli-darwin-arm64"
      sha256 "c86fd96698d5baf16a104d158984790f0a10b1c9fcfac64475967c943ecf14e5"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.3/airvoice-cli-darwin-amd64"
      sha256 "8822f955472340348229da8ec2949e3aa012bb588515a98fbfa0eb434290b362"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.3.3/airvoice-cli-linux-arm64"
      sha256 "19ede288b3daca267729414578f1b7ffafd81ca89f63157966caf7e68de0f91d"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.3.3/airvoice-cli-linux-amd64"
      sha256 "1f93ef4654601220247c6374644f1c61dd6e84f4f546411562ecae26cfc97bd9"
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
