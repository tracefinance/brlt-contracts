#!/bin/bash

# source.sh - Go source file viewer by package
# Displays Go files in specified packages, excluding test files

# Check if at least one argument is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <package_path1> [package_path2] [package_path3] ..."
    echo "Example: $0 internal/keygen internal/wallet"
    exit 1
fi

# Get the current directory name (assuming it's the project root)
PROJECT_ROOT=$(basename "$(pwd)")

# Function to list files in flat format
list_files() {
    # Process each package
    for package in "$@"; do
        if [ -d "$package" ]; then
            # Find all .go files in this package (excluding test files)
            find "$package" -type f -name "*.go" -not -name "*_test.go" 2>/dev/null | sort | while read -r file; do
                echo "$PROJECT_ROOT/$file"
            done
        fi
    done
}

# Function to process files in a package
process_package() {
    local package_path=$1
    local files_found=()
    
    # Find all .go files in the package, excluding test files
    while read -r file; do
        if [ -n "$file" ]; then
            files_found+=("$file")
        fi
    done < <(find "$package_path" -type f -name "*.go" -not -name "*_test.go" 2>/dev/null)
    
    # Process each found file
    for file in "${files_found[@]}"; do
        # Extract package name from first line of file
        package_name=$(grep -m 1 "^package " "$file" | sed 's/package \([a-zA-Z0-9_]*\).*/\1/')
        
        echo "---"
        echo "$file (package: $package_name)"
        echo "---"
        cat "$file"
        echo ""
        echo ""
    done
}

# Check if all provided packages exist
valid_packages=()
for package in "$@"; do
    if [ -d "$package" ]; then
        valid_packages+=("$package")
    else
        echo "Warning: Package directory '$package' does not exist or is not a directory."
    fi
done

# Print the flat file list for all valid packages
echo "Files:"
list_files "${valid_packages[@]}"
echo ""

# Process each valid package
for package in "${valid_packages[@]}"; do
    process_package "$package"
done
