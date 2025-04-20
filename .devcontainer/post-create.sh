#!/bin/bash
set -e

# Add task-master alias
echo "Setting up task-master alias..."
echo 'alias tm="task-master"' >> ~/.zshrc

# Install Go tools
echo "Installing Go tools..."
go install golang.org/x/tools/gopls@latest

# Install project dependencies
echo "Installing project dependencies..."
make deps

echo "Post-create setup completed successfully" 