#!/bin/bash

# OED MCP Server Installation Script

set -e

echo "OED MCP Server Installation"
echo "==========================="
echo ""

# Check for Go
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.21 or later."
    echo "Visit: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Found Go version: $GO_VERSION"
echo ""

# Build the server
echo "Building OED MCP server..."
go build -o oedmcp main.go

if [ $? -eq 0 ]; then
    echo "✓ Build successful!"
else
    echo "✗ Build failed. Please check error messages above."
    exit 1
fi

echo ""
echo "Configuration"
echo "-------------"

# Check for existing config
if [ -f "oed_config.json" ]; then
    echo "✓ Configuration file found: oed_config.json"
else
    echo "No configuration file found."
    echo ""
    echo "Please configure the OED data paths. You can either:"
    echo ""
    echo "1. Set environment variables:"
    echo "   export OED_DATA_PATH=/path/to/oed2"
    echo "   export OED_INDEX_PATH=/path/to/oed2index"
    echo ""
    echo "2. Create a configuration file:"
    echo "   cp oed_config.example.json oed_config.json"
    echo "   # Then edit oed_config.json with your paths"
    echo ""

    read -p "Would you like to create a config file now? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Enter path to oed2 data file: " data_path
        read -p "Enter path to oed2index file: " index_path

        cat > oed_config.json <<EOF
{
  "data_path": "$data_path",
  "index_path": "$index_path"
}
EOF
        echo "✓ Configuration file created: oed_config.json"
    fi
fi

echo ""
echo "Claude Desktop Integration"
echo "--------------------------"

# Detect OS and show appropriate config path
if [[ "$OSTYPE" == "darwin"* ]]; then
    CONFIG_PATH="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    CONFIG_PATH="$HOME/.config/Claude/claude_desktop_config.json"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ "$OSTYPE" == "win32" ]]; then
    CONFIG_PATH="%APPDATA%\\Claude\\claude_desktop_config.json"
else
    CONFIG_PATH="Unknown - check Claude Desktop documentation"
fi

echo "To use with Claude Desktop, add this to your config:"
echo "Location: $CONFIG_PATH"
echo ""
echo '{
  "mcpServers": {
    "oed": {
      "command": "'$(pwd)'/oedmcp"
    }
  }
}'
echo ""

# Test the installation
echo "Testing Installation"
echo "-------------------"

if [ -f "oed_config.json" ] || [ ! -z "$OED_DATA_PATH" ]; then
    echo -n "Testing server initialization... "

    OUTPUT=$(echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./oedmcp 2>&1 | head -1)

    if echo "$OUTPUT" | grep -q '"result"'; then
        echo "✓ Success!"
        echo ""
        echo "Installation complete! Restart Claude Desktop to use the OED tools."
    else
        echo "✗ Server test failed"
        echo "Please check your configuration and data file paths."
        exit 1
    fi
else
    echo "Skipping test (no configuration found)"
    echo ""
    echo "Installation complete. Please configure data paths before use."
fi

echo ""
echo "For more information, see README.md"