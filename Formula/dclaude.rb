class Dclaude < Formula
  desc "Containerized Claude Code runner with Docker isolation"
  homepage "https://github.com/jedi4ever/dclaude"
  url "https://github.com/jedi4ever/dclaude/archive/refs/tags/v1.1.0.tar.gz"
  sha256 "REPLACE_WITH_ACTUAL_SHA256"
  license "MIT"
  version "1.1.0"

  depends_on "docker"

  def install
    bin.install "dist/dclaude-standalone.sh" => "dclaude"
  end

  test do
    assert_match "1.1.0", shell_output("#{bin}/dclaude --version 2>&1", 0)
  end

  def caveats
    <<~EOS
      dclaude requires Docker to be running.

      To get started:
        1. Ensure Docker Desktop is running
        2. Run: dclaude --version

      For authentication, you can either:
        - Run 'claude login' (uses ~/.claude config)
        - Set ANTHROPIC_API_KEY environment variable

      Documentation: https://github.com/jedi4ever/dclaude
    EOS
  end
end
