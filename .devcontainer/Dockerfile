FROM mcr.microsoft.com/devcontainers/go:1.23-bullseye

# Install any additional OS packages needed (e.g., sqlite3)
# The base image includes git, curl, wget, make, sudo, etc.
RUN sudo apt-get update && sudo apt-get install -y --no-install-recommends sqlite3 jq \
    && sudo apt-get clean -y \
    && sudo rm -rf /var/lib/apt/lists/*

# Install Node.js and npm using the official Node.js distribution
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash - \
    && sudo apt-get install -y nodejs \
    && sudo apt-get clean -y \
    && sudo rm -rf /var/lib/apt/lists/*

# Enable CGO for C interoperability; required by dependencies like github.com/mattn/go-sqlite3
ENV CGO_ENABLED=1

# Set the working directory
WORKDIR /workspaces/vault0

# Install Go development tools
# These will be installed into the user's GOBIN, typically accessible in PATH
# Running as the default 'vscode' user provided by the base image
USER vscode
RUN go install github.com/google/wire/cmd/wire@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest

# Install task-master-ai globally with sudo
RUN sudo npm install -g task-master-ai

# The container will start as 'vscode' user by default from the base image
