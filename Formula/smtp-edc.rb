class SmtpEdc < Formula
  desc "A powerful, cross-platform SMTP testing tool written in Go"
  homepage "https://github.com/asachs01/smtp-edc"
  url "https://github.com/asachs01/smtp-edc/releases/download/latest/smtp-edc_Darwin_x86_64.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"  # This will be updated by GoReleaser
  license "MIT"

  depends_on "go" => :build

  def install
    bin.install "smtp-edc"
  end

  test do
    assert_match "smtp-edc", shell_output("#{bin}/smtp-edc --version")
  end
end
