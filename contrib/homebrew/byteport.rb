# BytePort CLI Homebrew Formula
# Installation:
#   brew tap kooshapari/byteport
#   brew install byteport
#
# This file is hosted at:
#   https://github.com/kooshapari/homebrew-byteport/blob/main/byteport.rb
#
# When updating, increment `version` and `revision` (only if formula change).

class Byteport < Formula
  desc "Multi-cloud deployment CLI for BytePort — ship to AWS, GCP, Azure, Fly, or local Docker"
  homepage "https://github.com/kooshapari/BytePort"
  version "0.1.0"
  license "Apache-2.0"

  # SHA256 pinning is mandatory for reproducible installs.
  # Run `./scripts/compute-formula-sha.sh` after each release to update.
  sha256 "REPLACE_WITH_RELEASE_SHA256"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/kooshapari/BytePort/releases/download/v#{version}/byteport-cli-v#{version}-aarch64-apple-darwin.tar.xz"
    else
      url "https://github.com/kooshapari/BytePort/releases/download/v#{version}/byteport-cli-v#{version}-x86_64-apple-darwin.tar.xz"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/kooshapari/BytePort/releases/download/v#{version}/byteport-cli-v#{version}-aarch64-unknown-linux-gnu.tar.xz"
    else
      url "https://github.com/kooshapari/BytePort/releases/download/v#{version}/byteport-cli-v#{version}-x86_64-unknown-linux-gnu.tar.xz"
    end
  end

  depends_on "openssl@3"
  depends_on "pkg-config"

  def install
    bin.install "byteport"
    zsh_completion.install "completions/_byteport"
    fish_completion.install "completions/byteport.fish"
    bash_completion.install "completions/byteport.bash"
  end

  test do
    # Smoke test: --version must succeed
    assert_match version.to_s, shell_output("#{bin}/byteport --version")
    assert_match "deploy", shell_output("#{bin}/byteport --help")
  end
end
