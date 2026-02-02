ARG NODE_VERSION=20
FROM node:${NODE_VERSION}-slim

# Build arguments for user ID and group ID
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG USERNAME=claude
ARG CLAUDE_VERSION=latest
ARG GO_VERSION=1.23
ARG UV_VERSION=0.5.11
ARG DCLAUDE_EXTENSIONS=claude

# Install dependencies, GitHub CLI, Docker CLI, and Docker daemon (for DinD)
RUN apt-get update && apt-get install -y \
    curl \
    gnupg \
    git \
    sudo \
    ripgrep \
    ca-certificates \
    iptables \
    ipset \
    dnsutils \
    supervisor \
    && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | gpg --dearmor -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian bookworm stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y gh docker-ce-cli docker-ce containerd.io \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "amd64" ]; then GO_ARCH="amd64"; \
    elif [ "$ARCH" = "arm64" ]; then GO_ARCH="arm64"; \
    else echo "Unsupported architecture: $ARCH" && exit 1; fi && \
    if [ "${GO_VERSION}" = "latest" ]; then \
        ACTUAL_GO_VERSION=$(curl -s 'https://go.dev/VERSION?m=text' | head -1); \
    else \
        ACTUAL_GO_VERSION="go${GO_VERSION}"; \
    fi && \
    curl -fsSL "https://go.dev/dl/${ACTUAL_GO_VERSION}.linux-${GO_ARCH}.tar.gz" -o /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz

# Add Go to PATH
ENV PATH="/usr/local/go/bin:${PATH}"

# Create user with matching UID/GID from host
# Note: GID 20 may already exist (staff/dialout group), so we just add user to that group
RUN (groupadd -g ${GROUP_ID} ${USERNAME} 2>/dev/null || true) \
    && useradd -m -u ${USER_ID} -g ${GROUP_ID} -s /bin/bash ${USERNAME} 2>/dev/null || usermod -u ${USER_ID} ${USERNAME} \
    && echo "${USERNAME} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Install UV (Python package manager) as the user using official install script
# This enables 'uv self update' to work inside containers
USER ${USERNAME}
RUN if [ "${UV_VERSION}" = "latest" ]; then \
        curl -LsSf https://astral.sh/uv/install.sh | sh; \
    else \
        curl -LsSf https://astral.sh/uv/${UV_VERSION}/install.sh | sh; \
    fi

USER root

# Copy extensions and make scripts executable
COPY extensions/ /usr/local/share/dclaude/extensions/
RUN chmod +x /usr/local/share/dclaude/extensions/install.sh && \
    find /usr/local/share/dclaude/extensions -name "*.sh" -exec chmod +x {} \;

# Install requested extensions (as claude user for go installs)
# Pass CLAUDE_VERSION so the claude extension can use it
USER ${USERNAME}
RUN CLAUDE_VERSION="${CLAUDE_VERSION}" /usr/local/share/dclaude/extensions/install.sh "${DCLAUDE_EXTENSIONS}"

# Add Go, UV, and extensions to PATH environment and bash profile
# UV installs to ~/.local/bin, Go tools install to ~/go/bin by default
ENV PATH="/home/${USERNAME}/go/bin:/home/${USERNAME}/.local/bin:/usr/local/go/bin:${PATH}"
RUN echo 'export PATH="/home/${USERNAME}/go/bin:/home/${USERNAME}/.local/bin:/usr/local/go/bin:$PATH"' >> /home/${USERNAME}/.bashrc

# Set working directory
WORKDIR /workspace

# Change ownership of workspace (use numeric IDs to avoid group name issues)
USER root
RUN chown -R ${USER_ID}:${GROUP_ID} /workspace

# Copy entrypoint wrapper for DinD support
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Copy firewall initialization script
COPY init-firewall.sh /usr/local/bin/init-firewall.sh
RUN chmod +x /usr/local/bin/init-firewall.sh

# Switch to non-root user
USER ${USERNAME}

# Add version labels for tracking
LABEL org.opencontainers.image.title="dclaude"
LABEL org.opencontainers.image.description="Containerized environment for AI coding agents"
LABEL org.opencontainers.image.authors="https://github.com/anthropics/claude-code"

# Tool version labels (populated at build time)
LABEL tools.git="installed"
LABEL tools.gh="installed"
LABEL tools.ripgrep="installed"
LABEL tools.node="installed"
LABEL tools.go="installed"
LABEL tools.uv="installed"

# Entry point will be our wrapper script (handles DinD mode)
# Empty CMD means interactive session by default
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD []
