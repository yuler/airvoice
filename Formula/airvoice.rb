class Airvoice < Formula
  desc "Voice-to-text CLI for Airvoice"
  homepage "https://github.com/yuler/airvoice"
  version "0.2.4"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.2.4/airvoice-cli-darwin-arm64"
      sha256 "05ca2c0566e4fafe3bbadb3d8ed892a139ede320e8378b8d1d89094593e3f5ff"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.2.4/airvoice-cli-darwin-amd64"
      sha256 "108c9195e3c77561337fddf9910901dae5999cc82242c775f698c7c7e8711950"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/yuler/airvoice/releases/download/v0.2.4/airvoice-cli-linux-arm64"
      sha256 "186ad56c6fc1ceb8e125c0135fd75913e7b2095d872508cda01f9f416653c92e"
    else
      url "https://github.com/yuler/airvoice/releases/download/v0.2.4/airvoice-cli-linux-amd64"
      sha256 "30832b21d2b75b9b701158701cbadbcae586723cba28c91b5284fdba2816fae4"
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
