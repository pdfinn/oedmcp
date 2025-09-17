# Deployment Guide for GitHub

This document outlines how to prepare the OED MCP Server for public release on GitHub while respecting Oxford University Press's intellectual property rights.

## Pre-Deployment Checklist

### ✅ Legal Compliance
- [ ] No OED data files in repository
- [ ] Clear disclaimers about OED ownership
- [ ] MIT license for code only (not data)
- [ ] Instructions require users to obtain their own OED data

### ✅ Configuration
- [ ] Data paths are configurable (not hardcoded)
- [ ] Multiple configuration methods (env vars, config files)
- [ ] Example configuration provided
- [ ] Local config files in .gitignore

### ✅ Documentation
- [ ] Clear README with legal notice
- [ ] Installation instructions
- [ ] Configuration guide
- [ ] Usage examples
- [ ] Contributing guidelines
- [ ] License file

### ✅ Code Quality
- [ ] No hardcoded paths
- [ ] Error messages guide users
- [ ] Clean separation of data access and server logic
- [ ] Proper error handling

## Repository Structure

```
oedmcp/
├── .gitignore                 # Excludes data and local configs
├── LICENSE                    # MIT with OED disclaimer
├── README.md                  # Main documentation
├── CONTRIBUTING.md            # Contribution guidelines
├── DEPLOYMENT.md              # This file
├── go.mod                     # Go module definition
├── go.sum                     # Go dependencies
├── main.go                    # MCP server implementation
├── install.sh                 # Installation helper
├── build.sh                   # Build script
├── oed_config.example.json    # Example configuration
├── config/
│   └── config.go             # Configuration management
└── dict/
    └── oed.go                # Dictionary access layer
```

## GitHub Release Steps

1. **Clean the Repository**
   ```bash
   # Remove any local configs or test files
   rm -f oed_config.json
   rm -f test_*.sh
   rm -f *.log
   ```

2. **Verify .gitignore**
   ```bash
   # Ensure no OED data can be committed
   git status --ignored
   ```

3. **Test Build**
   ```bash
   go build -o oedmcp
   ./oedmcp  # Should show config error
   ```

4. **Create GitHub Repository**
   - Set repository to public
   - Add description: "MCP server for Oxford English Dictionary access (requires OED data files)"
   - Add topics: `mcp`, `dictionary`, `oed`, `claude`, `ai-tools`

5. **Initial Commit**
   ```bash
   git init
   git add .
   git commit -m "Initial release of OED MCP Server"
   git branch -M main
   git remote add origin https://github.com/yourusername/oedmcp.git
   git push -u origin main
   ```

6. **Create Release**
   - Tag: v1.0.0
   - Title: "OED MCP Server v1.0.0"
   - Description: Include legal notice and basic usage

## Important Reminders

### Never Include
- OED data files (oed2, oed2index)
- Screenshots with extensive OED content
- Configuration files with actual paths
- Any Oxford University Press copyrighted material

### Always Include
- Legal disclaimers
- Requirement for users to obtain OED legally
- Configuration instructions
- Clear statement that this is interface software only

## Support Channels

- GitHub Issues for bug reports
- GitHub Discussions for feature requests
- Pull requests for contributions

## License Summary

The MIT License applies to:
- All Go source code
- Scripts and configuration examples
- Documentation written for this project

The MIT License does NOT apply to:
- OED data files
- OED content
- Oxford University Press materials

## Final Notes

This project demonstrates:
1. Responsible handling of proprietary data
2. Clean separation of interface and data
3. Respect for intellectual property
4. Value-add software that requires legal data access

By releasing this way, we:
- Enable legitimate OED users to enhance their access
- Respect Oxford University Press's rights
- Provide educational value about MCP servers
- Demonstrate proper software architecture