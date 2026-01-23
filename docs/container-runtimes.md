# Container Runtimes

cc-sandbox supports multiple container runtimes and automatically detects the best configuration for your system.

## Supported Runtimes

| Runtime           | Platforms             | Notes                                      |
|-------------------|-----------------------|--------------------------------------------|
| Docker Desktop    | Windows, macOS, Linux | Full support, recommended                  |
| OrbStack          | macOS                 | Automatic detection, recommended for macOS |
| Podman            | Linux                 | Rootless by default                        |
| Docker (rootless) | Linux                 | Automatic detection                        |

## Docker Desktop

### Windows

Docker Desktop for Windows is fully supported. Ensure WSL 2 backend is enabled for best performance.

```bash
cc-sandbox claude
```

### macOS

Docker Desktop for macOS works out of the box.

```bash
cc-sandbox claude
```

### Linux

Docker Desktop for Linux is supported. The CLI automatically detects whether you're using regular Docker or Docker
Desktop.

## OrbStack (macOS)

[OrbStack](https://orbstack.dev/) is automatically detected when installed on macOS. It provides faster container
startup and lower resource usage.

cc-sandbox detects OrbStack by:

1. Checking for the OrbStack socket at `~/.orbstack/run/docker.sock`
2. Checking Docker info for "OrbStack" in the operating system field

No special configuration needed:

```bash
cc-sandbox claude
```

## Podman (Linux)

Podman is supported with automatic UID mapping via `--userns=keep-id`.

### Automatic Detection

If Podman is available and Docker is not, cc-sandbox automatically uses Podman:

```bash
cc-sandbox claude  # Uses Podman if Docker isn't available
```

### Explicit Selection

```bash
cc-sandbox --runtime podman claude
```

Or via environment variable:

```bash
export CC_SANDBOX_RUNTIME=podman
cc-sandbox claude
```

### Podman Notes

- Uses `--userns=keep-id` for proper UID mapping
- Permission mode is set to `skip` (uses `--dangerously-skip-permissions` in Claude)
- Works with rootless Podman by default

## Rootless Docker

cc-sandbox automatically detects rootless Docker and adjusts its behavior.

### Detection Methods

1. **Docker info** - Checks for "rootless" in security options
2. **Socket location** - Checks for socket in user's home directory:
    - `~/.docker/run/docker.sock`
    - `~/.local/share/docker/run/docker.sock`

### Behavior in Rootless Mode

When rootless Docker is detected:

- Container runs as root (`-u 0:0`)
- Uses `--userns=host` for proper namespace mapping
- Permission mode is set to `accept` (accepts Claude's permission prompts)

### Override Detection

```bash
# Force root mode on
cc-sandbox --root=true claude

# Force root mode off
cc-sandbox --root=false claude

# Via environment variable
CC_SANDBOX_ROOT=true cc-sandbox claude
```

## Docker Socket Mounting

When using `--docker`, the CLI:

1. Mounts the Docker socket into the container
2. Adds the socket's group ID via `--group-add`
3. On macOS, adds group 0 for Docker Desktop/OrbStack compatibility

### Socket Locations

The CLI checks these locations in order:

1. `/var/run/docker.sock`
2. `~/.orbstack/run/docker.sock` (OrbStack)
3. `~/.docker/run/docker.sock` (rootless Docker)

Override with environment variable:

```bash
CC_SANDBOX_DOCKER_SOCKET=/custom/docker.sock cc-sandbox --docker claude
```

## Root Mode Behavior

The `--root` flag controls whether the container runs as root:

| Value            | Behavior                                     |
|------------------|----------------------------------------------|
| `auto` (default) | Root for rootless Docker, non-root otherwise |
| `true`           | Always run as root                           |
| `false`          | Always run as non-root (claude user)         |

### When to Use Root Mode

- **Rootless Docker** - Automatic, required for proper operation
- **Permission issues** - If you encounter file permission problems
- **System-level operations** - When the container needs elevated access

### Security Considerations

Running as root is safe because:

- The container is isolated from the host
- With rootless Docker, root in the container is your user on the host
- No sudo is installed in the container

## Host Network Mode

The `--host-network` flag enables host network mode:

```bash
cc-sandbox --host-network claude
```

### Use Cases

- Accessing services on `localhost` from inside the container
- Docker-in-Docker with port mappings that need localhost access
- Development servers running on the host

### Example

```bash
# Start a service on host
npm run dev  # Runs on localhost:3000

# In another terminal, access it from container
cc-sandbox --host-network claude -p "test the API at localhost:3000"
```

## Troubleshooting

### Docker Not Found

```bash
# Check Docker is installed and running
docker --version
docker info

# Check socket permissions
ls -la /var/run/docker.sock
```

### Podman Permission Errors

```bash
# Ensure Podman is running rootless
podman info | grep rootless

# Reset Podman if needed
podman system reset
```

### OrbStack Not Detected

```bash
# Verify OrbStack socket exists
ls -la ~/.orbstack/run/docker.sock

# Check Docker info
docker info | grep -i orbstack
```

### Debug Mode

Enable debug output to see runtime detection:

```bash
CC_SANDBOX_DEBUG=1 cc-sandbox claude
```
