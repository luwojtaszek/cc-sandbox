# cc-sandbox

**cc-sandbox** is a CLI tool for running Claude Code in isolated Docker containers with proper UID/GID mapping, optional credential mounting, and Docker socket access.

**DO AFTER EVERY IMPLEMENTATION**
- `./scripts/agent_checks.sh` - verification script, fix all reported problems

## Project Structure

- `cli/` - Go CLI source code
- `docker/` - Docker image definitions (base, docker, bun-full)
- `scripts/` - Build and check scripts

## Development

```bash
# Build CLI
make cli

# Run all checks
make check

# Individual checks
make build    # Compile
make test     # Run tests
make lint     # Run linter
make fmt      # Format code
```

## Docker Images

- `cc-sandbox:base` - Base image with Node.js, Claude Code
- `cc-sandbox:docker` - Adds Docker-in-Docker support
- `cc-sandbox:bun-full` - Adds Bun runtime
