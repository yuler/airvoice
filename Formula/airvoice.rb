class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.4.9"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.9/airvoice-cli-darwin-arm64"
      sha256 "0f8a631b8580f4b721ebc1b88adc62dcf5510aef3a0dc13d46ac8561c694dc3c"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.9/airvoice-cli-darwin-amd64"
      sha256 "9429938b188abfbce970a49813eec157521122e0d5e95bea23a7e0d125a3eddb"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.4.9/airvoice-cli-linux-arm64"
      sha256 "3b6b416d9027fa22b185226cf7e6c07674aede81a8d2de622e38ce76149de54f"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.4.9/airvoice-cli-linux-amd64"
      sha256 "48089bfb510bea45dd25ab7ab9d0afec3cf3c462f49a20f11e6c60505466b563"
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
