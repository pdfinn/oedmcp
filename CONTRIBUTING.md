# Contributing to OED MCP Server

Thank you for your interest in contributing to the OED MCP Server project! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive criticism
- Help maintain a welcoming environment for all contributors

## Legal Considerations

**CRITICAL**: Never commit or share OED data files. These are proprietary and copyrighted by Oxford University Press.

- Do not include any dictionary data in commits
- Keep all data paths configurable
- Test with your own legally obtained OED files
- Do not share screenshots containing extensive OED content

## How to Contribute

### Reporting Issues

1. Check existing issues to avoid duplicates
2. Provide clear description of the problem
3. Include steps to reproduce
4. Specify your environment (OS, Go version, etc.)
5. Do NOT include OED content in issue reports

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Write or update tests as needed
5. Ensure all tests pass: `go test ./...`
6. Commit with clear messages: `git commit -m "Add feature: description"`
7. Push to your fork: `git push origin feature/your-feature`
8. Open a Pull Request

### Development Setup

1. Fork and clone the repository
2. Install Go 1.21+
3. Get dependencies: `go mod download`
4. Set up your config file (see README.md)
5. Build: `go build -o oedmcp`

## Development Guidelines

### Code Style

- Follow standard Go formatting: `go fmt ./...`
- Use meaningful variable names
- Add comments for complex logic
- Keep functions focused and small

### Testing

- Write tests for new features
- Maintain existing test coverage
- Mock OED data for tests (don't use real data)
- Run tests before submitting PR: `go test ./...`

### Documentation

- Update README.md for user-facing changes
- Add code comments for complex functions
- Update configuration examples if needed
- Document new MCP tools thoroughly

## Areas for Contribution

### High Priority

- [ ] Improved error handling and user feedback
- [ ] Performance optimizations for large lookups
- [ ] Better parsing of complex OED entries
- [ ] Support for additional dictionary formats

### Feature Ideas

- [ ] Caching layer for frequently looked up words
- [ ] Fuzzy search/spell correction
- [ ] Advanced search (by definition, quotation date, etc.)
- [ ] Export tools (save lookups to file)
- [ ] Statistics and analytics tools

### Infrastructure

- [ ] Automated testing pipeline
- [ ] Docker containerization
- [ ] Cross-platform build scripts
- [ ] Installation packages for various OS

## Testing Your Changes

### Unit Tests

```bash
go test ./dict
go test ./config
```

### Integration Tests

```bash
# Test MCP protocol compliance
./scripts/test_mcp_protocol.sh

# Test all tools
./scripts/test_all_tools.sh
```

### Manual Testing

Test each tool with your Claude Desktop setup:
1. Build your changes
2. Restart Claude Desktop
3. Test affected tools
4. Verify error handling

## Commit Message Guidelines

Use clear, descriptive commit messages:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `perf:` Performance improvements
- `chore:` Maintenance tasks

Examples:
```
feat: add wildcard search support
fix: correct etymology parsing for compound words
docs: update installation instructions for Windows
```

## Questions?

Feel free to open an issue for:
- Clarification on implementation
- Discussion of new features
- Help with development setup

## License

By contributing, you agree that your contributions will be licensed under the MIT License.