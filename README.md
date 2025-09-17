# OED MCP Server

An MCP (Model Context Protocol) server that provides AI assistants with access to the Oxford English Dictionary 2nd Edition. This allows Claude and other MCP-compatible AI assistants to look up word definitions, etymologies, and historical usage directly from the OED.

## ⚠️ Important Legal Notice

**This software does not include any OED data.** The Oxford English Dictionary is proprietary content owned by Oxford University Press. You must have legal access to OED2 data files to use this server. This project only provides the interface to read properly licensed OED data that you provide.

## Features

- 🔍 **Word Lookup**: Full dictionary entries with pronunciations, definitions, and quotations
- 📚 **Etymology Search**: Detailed word origins and historical development
- 🔤 **Prefix Search**: Find words starting with specific prefixes
- 🎲 **Random Word**: Discover random dictionary entries
- 📦 **Batch Lookup**: Look up multiple words simultaneously

## Prerequisites

- Go 1.21 or later
- Legal access to OED2 data files in Plan 9 format:
  - `oed2` - Main data file (~520 MB)
  - `oed2index` - Index file (~5.6 MB)
- Claude Desktop or another MCP-compatible client

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/oedmcp.git
cd oedmcp
```

### 2. Configure Data Paths

The server needs to know where your OED data files are located. You can configure this in several ways (in order of priority):

#### Option A: Environment Variables
```bash
export OED_DATA_PATH=/path/to/your/oed2
export OED_INDEX_PATH=/path/to/your/oed2index
```

#### Option B: Configuration File
Copy the example configuration and edit it:
```bash
cp oed_config.example.json oed_config.json
```

Edit `oed_config.json`:
```json
{
  "data_path": "/path/to/your/oed2",
  "index_path": "/path/to/your/oed2index"
}
```

#### Option C: User Home Directory
Create a configuration in your home directory:
```bash
mkdir -p ~/.oed_mcp
cp oed_config.example.json ~/.oed_mcp/config.json
# Edit ~/.oed_mcp/config.json with your paths
```

### 3. Build the Server

```bash
go build -o oedmcp
```

Or use the provided build script:
```bash
./build.sh
```

### 4. Configure Claude Desktop

Add the server to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

Add this to your `mcpServers` section:

```json
{
  "mcpServers": {
    "oed": {
      "command": "/absolute/path/to/oedmcp",
      "env": {
        "OED_DATA_PATH": "/path/to/your/oed2",
        "OED_INDEX_PATH": "/path/to/your/oed2index"
      }
    }
  }
}
```

### 5. Restart Claude Desktop

The OED tools will be available after restarting Claude Desktop.

## Usage

Once configured, you can ask Claude to look up words:

- "Look up 'serendipity' in the OED"
- "What's the etymology of 'computer'?"
- "Search for words starting with 'astro'"
- "Give me a random word from the OED"
- "Look up 'cat', 'dog', and 'bird' in the dictionary"

## Available MCP Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `oed_lookup` | Look up a complete dictionary entry | `word` (required), `include_etymology` (optional) |
| `oed_etymology` | Get detailed etymology information | `word` (required) |
| `oed_search` | Search for words by prefix | `prefix` (required), `limit` (optional, max 50) |
| `oed_random` | Get a random dictionary entry | None |
| `oed_multi_lookup` | Look up multiple words at once | `words` (required, comma-separated) |

## Data Format

This server works with OED2 data in Plan 9 dictionary format. The data files should be:

- **oed2**: Binary data file containing dictionary entries with embedded XML-style tags
- **oed2index**: Tab-separated index file with format: `word[TAB]offset`

The server parses OED2's tag format including:
- `<hw>...</hw>` - Headwords
- `<etym>...</etym>` - Etymology
- `<pr>...</pr>` - Pronunciation
- Various other semantic and formatting tags

## Development

### Project Structure
```
oedmcp/
├── main.go              # MCP server implementation
├── dict/
│   └── oed.go          # Dictionary access layer
├── config/
│   └── config.go       # Configuration management
├── go.mod              # Go module definition
└── go.sum              # Go module checksums
```

### Testing

Create a test configuration:
```bash
cp oed_config.example.json oed_config.json
# Edit with your data paths
```

Run tests:
```bash
go test ./...
```

Test the server directly:
```bash
# Initialize and test
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"0.1.0","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | ./oedmcp
```

## Troubleshooting

### Server won't start
- Verify OED data files exist at configured paths
- Check file permissions (must be readable)
- Ensure paths are absolute, not relative

### No results from lookups
- Verify the index file matches the data file
- Check that both files are from the same OED2 version
- Try a known common word like "dictionary" or "cat"

### Etymology not showing
- Ensure you're using `include_etymology: true` in lookup
- Some entries may not have etymology information

## Contributing

Contributions are welcome! Please note:
- Do NOT commit any OED data files
- Keep data paths configurable
- Maintain compatibility with MCP protocol
- Add tests for new features

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**Important**: This license applies only to the server code, not to any OED data. The OED remains the intellectual property of Oxford University Press.

## Acknowledgments

- Based on Plan 9's dictionary format and tools
- Uses the [mcp-go](https://github.com/mark3labs/mcp-go) library
- Inspired by the original Plan 9 `dict` implementation

## Disclaimer

This project is not affiliated with, endorsed by, or connected to Oxford University Press or the Oxford English Dictionary. Users must ensure they have appropriate rights to use OED data files with this software.