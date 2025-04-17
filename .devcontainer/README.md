# Dev Container for Vault0

This dev container configuration sets up a development environment with:

- Go 1.23
- Node.js LTS
- Essential development tools and utilities

## Features

- Pre-installed Go with CGO support
- Node.js LTS with npm
- SQLite3 for local database
- VS Code extensions for Go and JavaScript/TypeScript development
- Exposed ports:
  - 8080 (Go applications)
  - 3000 (Node.js applications)

## Usage

1. Install [VS Code](https://code.visualstudio.com/) and the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. Open this repository in VS Code
3. When prompted, select "Reopen in Container"
4. Wait for the container to build and initialize

## Customization

- Edit `devcontainer.json` to modify VS Code settings or add extensions
- Edit `Dockerfile` to install additional packages or modify environment 