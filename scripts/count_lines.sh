#!/bin/bash

# count_lines.sh - Script to count lines of code in vault0 project
# 
# This script counts lines of code in:
# - UI (React/TypeScript)
# - Backend (Golang)
# - Smart Contracts (Solidity)
# Both source and test files are counted

# Directory paths
UI_DIR="./ui"
CONTRACTS_DIR="./contracts"

# Default options
COUNT_UI=true
COUNT_BACKEND=true
COUNT_CONTRACTS=true
COUNT_SRC=true
COUNT_TESTS=true
VERBOSE=true

# Display usage
usage() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  -h, --help         Display this help message"
  echo "  -u, --ui-only      Count only UI code"
  echo "  -b, --backend-only Count only backend code"
  echo "  -c, --contracts-only Count only smart contracts"
  echo "  -s, --source-only  Count only source code (no tests)"
  echo "  -t, --tests-only   Count only test code (no source)"
  echo "  -q, --quiet        Display only the summary"
  exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      ;;
    -u|--ui-only)
      COUNT_UI=true
      COUNT_BACKEND=false
      COUNT_CONTRACTS=false
      shift
      ;;
    -b|--backend-only)
      COUNT_UI=false
      COUNT_BACKEND=true
      COUNT_CONTRACTS=false
      shift
      ;;
    -c|--contracts-only)
      COUNT_UI=false
      COUNT_BACKEND=false
      COUNT_CONTRACTS=true
      shift
      ;;
    -s|--source-only)
      COUNT_SRC=true
      COUNT_TESTS=false
      shift
      ;;
    -t|--tests-only)
      COUNT_SRC=false
      COUNT_TESTS=true
      COUNT_UI=false # UI doesn't have tests
      shift
      ;;
    -q|--quiet)
      VERBOSE=false
      shift
      ;;
    *)
      echo "Unknown option: $1"
      usage
      ;;
  esac
done

# Function to count lines in files and show results
count_and_show() {
  local name=$1
  local count=$2
  
  if [ "$VERBOSE" = true ]; then
    echo "$name:"
    printf "    %5d total\n" "$count"
    echo ""
  fi
}

# Function to get percentage
get_percentage() {
  local count=$1
  local total=$2
  if [ "$total" -eq 0 ]; then
    echo "0.0"
    return
  fi
  echo "scale=1; $count * 100 / $total" | bc
}

# Variables to store line counts
UI=0
BACKEND=0
BACKEND_TEST=0
CONTRACT=0
CONTRACT_TEST=0

if [ "$VERBOSE" = true ]; then
  echo "Counting lines of code in the project..."
  echo "----------------------------------------"
fi

# Count UI lines (only if counting source code)
if [ "$COUNT_UI" = true ] && [ "$COUNT_SRC" = true ]; then
  # Count UI source files
  UI=$(find ${UI_DIR}/src -type f \( -name "*.tsx" -o -name "*.ts" -o -name "*.jsx" -o -name "*.js" -o -name "*.css" \) | grep -v "node_modules" | xargs wc -l 2>/dev/null | awk 'END {print $1}')
  count_and_show "UI Source Code" "$UI"
fi

# Count backend lines
if [ "$COUNT_BACKEND" = true ]; then
  # Count backend source code
  if [ "$COUNT_SRC" = true ]; then
    BACKEND=$(find cmd internal -type f -name "*.go" | grep -v "_test.go" | xargs wc -l 2>/dev/null | awk 'END {print $1}')
    count_and_show "Backend Source Code" "$BACKEND"
  fi
  
  # Count backend test code
  if [ "$COUNT_TESTS" = true ]; then
    BACKEND_TEST=$(find cmd internal -type f -name "*_test.go" | xargs wc -l 2>/dev/null | awk 'END {print $1}')
    count_and_show "Backend Test Code" "$BACKEND_TEST"
  fi
fi

# Count contract lines
if [ "$COUNT_CONTRACTS" = true ]; then
  # Count contract source code
  if [ "$COUNT_SRC" = true ]; then
    CONTRACT=$(find ${CONTRACTS_DIR} -type f -name "*.sol" | grep -v "node_modules" | xargs wc -l 2>/dev/null | awk 'END {print $1}')
    count_and_show "Solidity Contracts Source Code" "$CONTRACT"
  fi
  
  # Count contract test code
  if [ "$COUNT_TESTS" = true ]; then
    CONTRACT_TEST=$(find ${CONTRACTS_DIR}/test -type f -name "*.js" | grep -v "node_modules" | xargs cat 2>/dev/null | wc -l || echo 0)
    count_and_show "Solidity Contracts Test Code" "$CONTRACT_TEST"
  fi
fi

# Calculate total
TOTAL=$((UI + BACKEND + BACKEND_TEST + CONTRACT + CONTRACT_TEST))

# Skip summary if there are no lines to count
if [ "$TOTAL" -eq 0 ]; then
  echo "No lines of code counted based on the specified options."
  exit 0
fi

echo "Total Lines Summary:"
echo "-------------------"

# Display component summaries with consistent formatting
if [ "$UI" -ne 0 ]; then
  printf "%-20s %5d  (%5.1f%%)\n" "UI:" "$UI" "$(get_percentage $UI $TOTAL)"
fi

if [ "$BACKEND" -ne 0 ]; then
  printf "%-20s %5d  (%5.1f%%)\n" "Backend Source:" "$BACKEND" "$(get_percentage $BACKEND $TOTAL)"
fi

if [ "$BACKEND_TEST" -ne 0 ]; then
  printf "%-20s %5d  (%5.1f%%)\n" "Backend Tests:" "$BACKEND_TEST" "$(get_percentage $BACKEND_TEST $TOTAL)"
fi

if [ "$CONTRACT" -ne 0 ]; then
  printf "%-20s %5d  (%5.1f%%)\n" "Contracts Source:" "$CONTRACT" "$(get_percentage $CONTRACT $TOTAL)"
fi

if [ "$CONTRACT_TEST" -ne 0 ]; then
  printf "%-20s %5d  (%5.1f%%)\n" "Contracts Tests:" "$CONTRACT_TEST" "$(get_percentage $CONTRACT_TEST $TOTAL)"
fi

echo ""

# Calculate source vs test totals
SRC_TOTAL=$((UI + BACKEND + CONTRACT))
TEST_TOTAL=$((BACKEND_TEST + CONTRACT_TEST))

# Display source vs test summaries
if [ "$SRC_TOTAL" -ne 0 ]; then
  printf "%-20s %5d  (%5.1f%%)\n" "Source Code:" "$SRC_TOTAL" "$(get_percentage $SRC_TOTAL $TOTAL)"
fi

if [ "$TEST_TOTAL" -ne 0 ]; then
  printf "%-20s %5d  (%5.1f%%)\n" "Test Code:" "$TEST_TOTAL" "$(get_percentage $TEST_TOTAL $TOTAL)"
fi

echo "-------------------"
printf "%-20s %5d\n" "Grand Total:" "$TOTAL"