# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`linctl` is a comprehensive Go CLI tool for Linear's API, designed for both human users and AI agents. It's built using the Cobra CLI framework with a focus on structured output formats and comprehensive Linear API coverage.

## Commands Reference

### Build & Development
- `make build` - Build the binary locally
- `make deps` - Install/tidy dependencies
- `make test` - Run smoke tests (read-only commands only)
- `make lint` - Run golangci-lint (if installed)
- `make fmt` - Format code with go fmt
- `make install` - Install to /usr/local/bin (requires sudo)
- `make dev-install` - Create development symlink
- `make clean` - Clean build artifacts

### Testing
- `./smoke_test.sh` - Run all smoke tests
- `bash -x ./smoke_test.sh` - Run smoke tests with verbose output

### Running
- `go run main.go` - Run directly without building
- `./linctl` - Run built binary

## Code Architecture

### Project Structure
```
cmd/           - CLI command definitions (Cobra commands)
├── root.go    - Root command and global configuration
├── auth.go    - Authentication commands
├── issue.go   - Issue management commands
├── project.go - Project management commands
├── team.go    - Team management commands
├── user.go    - User management commands
├── comment.go - Comment commands
└── docs.go    - Documentation commands

pkg/           - Reusable packages
├── api/       - Linear API client and GraphQL queries
├── auth/      - Authentication utilities
├── output/    - Output formatting (table, JSON, plaintext)
└── utils/     - Utility functions (time parsing, etc.)

main.go        - Application entry point
```

### Key Architectural Patterns

**Command Structure**: Each command follows a consistent pattern:
- Uses Cobra for CLI structure
- Supports `--json`, `--plaintext` flags for output format
- Implements comprehensive error handling
- Uses Viper for configuration management

**API Client**: Centralized GraphQL client in `pkg/api/client.go`:
- Single HTTP client for all Linear API calls
- Structured GraphQL request/response handling
- Comprehensive error reporting

**Output Formatting**: Standardized output in `pkg/output/output.go`:
- Table output using `tablewriter`
- JSON output with proper marshaling
- Plaintext output for non-interactive use
- Color support via `fatih/color`

**Authentication**: Linear API key management in `pkg/auth/auth.go`:
- Config file storage (~/.linctl.yaml)
- Environment variable support
- Interactive authentication flow

### Key Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/olekukonko/tablewriter` - Table output formatting
- `github.com/fatih/color` - Colored terminal output

### Default Behavior
- Issue and project lists default to last 6 months for performance
- Issue lists exclude completed/canceled items by default
- All commands support JSON output for programmatic use
- Authentication stored in ~/.linctl.yaml

### Testing Strategy
- Smoke tests for all read-only commands in `smoke_test.sh`
- No unit tests currently implemented
- Manual testing recommended for write operations

### Version Management
- Version injected at build time via `-ldflags`
- Git-based versioning using tags or commit hash
- Embedded README.md in binary for `linctl docs` command

## Important Notes for Development

- Always use JSON output (`--json`) for programmatic access
- Read-only commands are safe for testing via smoke tests
- Write operations should be tested manually with caution
- Configuration file format is YAML stored in user home directory
- All Linear API calls use GraphQL endpoints