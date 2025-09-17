#!/bin/bash

# Build the OED MCP Server

echo "Building OED MCP Server..."
go build -o oedmcp .

if [ $? -eq 0 ]; then
    echo "Build successful! Binary created: ./oedmcp"
    echo ""
    echo "To use with Claude Desktop, add this to your config:"
    echo '  "oed": {'
    echo '    "command": "'$(pwd)'/oedmcp"'
    echo '  }'
else
    echo "Build failed!"
    exit 1
fi