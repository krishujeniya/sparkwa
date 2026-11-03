class Sparkwa < Formula
  desc "Headless WhatsApp to Google Gemini Spark Bridge"
  homepage "https://github.com/krishujeniya/sparkwa"
  version "1.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/krishujeniya/sparkwa/releases/download/v1.0.0/sparkwa-v1.0.0-darwin-arm64.tar.gz"
    else
      url "https://github.com/krishujeniya/sparkwa/releases/download/v1.0.0/sparkwa-v1.0.0-darwin-amd64.tar.gz"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/krishujeniya/sparkwa/releases/download/v1.0.0/sparkwa-v1.0.0-linux-arm64.tar.gz"
    else
      url "https://github.com/krishujeniya/sparkwa/releases/download/v1.0.0/sparkwa-v1.0.0-linux-amd64.tar.gz"
    end
  end

  def install
    bin.install "sparkwa"
  end

  service do
    run [opt_bin/"sparkwa", "run"]
    keep_alive true
    log_path var/"log/sparkwa.log"
    error_log_path var/"log/sparkwa.err"
  end

  test do
    assert_match "SparkWA", shell_output("#{bin}/sparkwa --help 2>&1", 0)
  end
end
