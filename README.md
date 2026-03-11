# Memex - Persistent Memory for Claude Code

Memex is a Model Context Protocol (MCP) server that provides persistent memory storage for AI assistants. It enables long-term memory capabilities through SQLite with full-text search (FTS5), allowing Claude Code to remember and recall information across sessions.

## Features

- **Persistent Memory**: Store and retrieve information across Claude Code sessions
- **Full-Text Search**: Fast BM25-ranked search using SQLite FTS5
- **Code Anchors**: Link memories to specific files, functions, and line ranges
- **Project Isolation**: Memories are automatically scoped to git repositories
- **Zero Configuration**: Works out of the box with sensible defaults
- **Lightweight**: Pure Go with SQLite - no external dependencies

## Architecture

```
Claude/AI Assistant
       │
       ▼
   MCP Client
       │
       ▼
  Memex Server (stdio)
       │
       ├─▶ Memory Service
       │      ├─ Create
       │      ├─ Get
       │      ├─ Search (BM25)
       │      ├─ List
       │      └─ Delete
       │
       └─▶ SQLite Storage (FTS5)
            ├─ memories table
            ├─ FTS5 search index
            └─ code_anchors table
```

## Troubleshooting

### Server not starting

**Check logs in stderr:**
```bash
memex 2>&1 | grep -i error
```

**Verify installation:**
```bash
which memex
memex --version
```

### Memories not persisting

**Check database path:**
```bash
# Default location
ls -l ~/.memex/memex.db

# Custom location
MEMEX_DATABASE_PATH=/tmp/test.db memex
```

**Verify permissions:**
```bash
chmod 700 ~/.memex
chmod 600 ~/.memex/memex.db
```

### Claude Code can't find server

**Verify settings.json syntax:**
```bash
cat ~/.claude/settings.json | jq .
```

**Check command path:**
```bash
which memex
# Should match the "command" in settings.json
```

**Restart Claude Code** after configuration changes.

### Search not working

**Verify FTS5 support:**
```bash
# Build requires CGO_ENABLED=1 and -tags "fts5"
make build
```

**Check database initialization:**
```bash
sqlite3 ~/.memex/memex.db "SELECT * FROM sqlite_master WHERE type='table';"
# Should show: memories, memories_fts, code_anchors
```

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**: All code must pass `make ci`
2. **Tests**: Add tests for new features
3. **Commits**: Use conventional commits (feat:, fix:, docs:, etc.)
4. **Documentation**: Update README for user-facing changes

### Development Workflow

```bash
# Fork and clone
git clone https://github.com/yourusername/memex
cd memex

# Create branch
git checkout -b feature/your-feature

# Make changes and test
make test
make lint

# Commit and push
git commit -m "feat: add your feature"
git push origin feature/your-feature

# Open PR
gh pr create
```

## Quick Start

### Installation

**Option 1: Using Go**
```bash
go install github.com/AnkushinDaniil/memex/cmd/memex@latest
```

**Option 2: Build from Source**
```bash
git clone https://github.com/AnkushinDaniil/memex
cd memex
make install
```

This installs the `memex` binary to `~/.local/bin/memex`. Make sure `~/.local/bin` is in your PATH:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Configuration

Add Memex to your Claude Code settings (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "memex": {
      "command": "memex",
      "args": [],
      "env": {
        "MEMEX_DATABASE_PATH": "~/.memex/memex.db",
        "MEMEX_LOG_LEVEL": "info"
      }
    }
  }
}
```

See [examples/claude-settings.json](examples/claude-settings.json) for more configuration options.

### Usage

Start Claude Code and use natural language to interact with your memories:

**Storing Information:**
```
You: "Remember that we use JWT for authentication with a 24-hour expiry"
Claude: *stores via memex_remember*
```

**Retrieving Information:**
```
You: "What did we decide about authentication?"
Claude: *searches via memex_recall* "We use JWT for authentication with a 24-hour expiry"
```

**Managing Memories:**
```
You: "List all memories about authentication"
Claude: *calls memex_list with tag filter*

You: "Forget the note about JWT expiry"
Claude: *calls memex_forget*
```

## MCP Tools

Memex provides 5 MCP tools:

### memex_remember

Store a new memory with optional code anchors, tags, and metadata.

**Parameters:**
- `content` (required): The content to remember
- `tags` (optional): Array of tags for categorization
- `priority` (optional): `low`, `normal`, or `high`
- `type` (optional): Memory type (bug-fix, gotcha, design-decision, etc.)
- `anchors` (optional): Array of code location anchors

**Example:**
```json
{
  "content": "We use JWT for authentication with 24-hour expiry",
  "tags": ["auth", "security"],
  "priority": "high",
  "type": "design-decision",
  "anchors": [
    {
      "file": "internal/auth/jwt.go",
      "function": "GenerateToken",
      "start_line": 45,
      "end_line": 60
    }
  ]
}
```

### memex_recall

Search memories using full-text search with BM25 ranking.

**Parameters:**
- `query` (required): Search query
- `limit` (optional): Maximum results (default: 10)
- `tags` (optional): Filter by tags
- `type` (optional): Filter by memory type

**Example:**
```json
{
  "query": "authentication JWT",
  "limit": 5,
  "tags": ["auth"]
}
```

### memex_forget

Delete a memory by ID.

**Parameters:**
- `memory_id` (required): ID of the memory to delete

### memex_list

List recent memories for the current project.

**Parameters:**
- `limit` (optional): Maximum results (default: 20)
- `tags` (optional): Filter by tags

### memex_stats

Get statistics about stored memories.

**Parameters:**
- `project_id` (optional): Project ID (defaults to current project)

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MEMEX_DATABASE_PATH` | `~/.memex/memex.db` | Database location |
| `MEMEX_LOG_LEVEL` | `info` | Log level: debug, info, warn, error |
| `MEMEX_MODE` | `local` | Mode: local or cloud |

### CLI Flags

```bash
memex --db /path/to/db.db       # Custom database path
memex --log-level debug          # Enable debug logging
memex --version                  # Print version
```

**Priority:** CLI flags > Environment variables > Defaults

## Project Detection

Memex automatically detects your project context:

1. **Git repository**: Uses the git root directory name as project ID
2. **Fallback**: Uses current working directory name

All memories are scoped to the detected project, ensuring isolation between different codebases.

## Development

### Build

```bash
make build       # Build binary to bin/memex
make install     # Install to ~/.local/bin
make uninstall   # Remove from ~/.local/bin
```

### Test

```bash
make test              # Run all tests
make test-coverage     # Run tests with coverage
make coverage-html     # Generate HTML coverage report
```

### Code Quality

```bash
make lint        # Run golangci-lint
make fmt         # Format code
make security    # Run security scans
make ci          # Full CI pipeline
```

### Manual Testing

```bash
# Test MCP protocol
./scripts/test-mcp.sh

# Integration test
./scripts/test-integration.sh
```

## Technology Stack

- **Language**: Go 1.26
- **Database**: SQLite with FTS5
- **Protocol**: MCP (Model Context Protocol)
- **Communication**: JSON-RPC 2.0 over stdio

## CI/CD

The project uses GitHub Actions for continuous integration:

- **Linting**: golangci-lint v2.10.1
- **Testing**: Unit tests with race detector
- **Security**: govulncheck + Trivy scanning
- **Coverage**: 80% threshold (warning)
- **Release**: Automated semantic versioning

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [Model Context Protocol](https://modelcontextprotocol.io/) - MCP specification
- [Claude Code](https://claude.ai/code) - AI coding assistant

## Support

- **Issues**: [GitHub Issues](https://github.com/AnkushinDaniil/memex/issues)
- **Discussions**: [GitHub Discussions](https://github.com/AnkushinDaniil/memex/discussions)
