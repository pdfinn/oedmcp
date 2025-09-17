#!/bin/bash

# Run tests with coverage report

echo "OED MCP Server - Test Suite"
echo "============================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Run tests with coverage
echo "Running unit tests..."
echo "--------------------"

# Test config package
echo -e "${YELLOW}Testing config package:${NC}"
go test -cover ./config
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Config tests passed${NC}"
else
    echo -e "${RED}✗ Config tests failed${NC}"
fi
echo ""

# Test dict package
echo -e "${YELLOW}Testing dict package:${NC}"
go test -cover ./dict
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Dict tests passed${NC}"
else
    echo -e "${RED}✗ Dict tests failed${NC}"
fi
echo ""

# Integration tests (only if config is available)
if [ -f "oed_config.json" ] || [ ! -z "$OED_DATA_PATH" ]; then
    echo -e "${YELLOW}Running integration tests:${NC}"
    go test -cover -short .
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Integration tests passed${NC}"
    else
        echo -e "${RED}✗ Integration tests failed${NC}"
    fi
else
    echo -e "${YELLOW}Skipping integration tests (no config found)${NC}"
fi
echo ""

# Overall coverage report
echo "Coverage Report"
echo "---------------"
go test -cover ./... 2>/dev/null | grep coverage || echo "Run with real OED data for full coverage"

echo ""
echo "Test Summary"
echo "------------"

# Count test files
test_files=$(find . -name "*_test.go" | wc -l | tr -d ' ')
echo "Test files: $test_files"

# Count test functions
test_functions=$(grep -h "^func Test" *_test.go **/*_test.go 2>/dev/null | wc -l | tr -d ' ')
echo "Test functions: $test_functions"

# List packages with tests
echo ""
echo "Packages with tests:"
for pkg in config dict; do
    if [ -f "$pkg/${pkg}_test.go" ]; then
        echo "  ✓ $pkg"
    fi
done
if [ -f "integration_test.go" ]; then
    echo "  ✓ integration"
fi

echo ""
echo "To run tests with verbose output: go test -v ./..."
echo "To generate coverage report: go test -coverprofile=coverage.out ./..."
echo "To view coverage in browser: go tool cover -html=coverage.out"