ARG NODE_VERSION=22
FROM node:${NODE_VERSION}-slim

# Build arguments
ARG CLAUDE_VERSION=latest

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    git \
    jq \
    sudo \
    ripgrep \
    ca-certificates \
    && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code globally (specific version or latest)
RUN if [ "$CLAUDE_VERSION" = "latest" ]; then \
        npm install -g @anthropic-ai/claude-code; \
    else \
        npm install -g @anthropic-ai/claude-code@$CLAUDE_VERSION; \
    fi

# Set working directory
WORKDIR /workspace

# Copy entrypoint script
COPY daytona-entrypoint.sh /usr/local/bin/daytona-entrypoint.sh
RUN chmod +x /usr/local/bin/daytona-entrypoint.sh

# Add version labels for tracking
LABEL org.opencontainers.image.title="dclaude-daytona"
LABEL org.opencontainers.image.description="Claude Code for Daytona with Git, GitHub CLI, and Ripgrep"

# Tool version labels
LABEL tools.claude="installed"
LABEL tools.git="installed"
LABEL tools.gh="installed"
LABEL tools.ripgrep="installed"
LABEL tools.node="installed"

# Entry point
ENTRYPOINT ["/usr/local/bin/daytona-entrypoint.sh"]
CMD []
