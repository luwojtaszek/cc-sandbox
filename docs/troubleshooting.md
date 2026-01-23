# Troubleshooting

Common issues and their solutions.

## Permission Issues

### Files Created with Wrong Ownership

**Symptom:** Files created in the container are owned by a different user on the host.

**Solution:** cc-sandbox uses fixuid for UID/GID mapping. Ensure you're running through the CLI:

```bash
# Correct - CLI handles UID mapping
cc-sandbox claude

# If running Docker directly, specify user
docker run --user $(id -u):$(id -g) ghcr.io/luwojtaszek/cc-sandbox:base
```

### Permission Denied on Workspace

**Symptom:** Cannot write to `/workspace` inside the container.

**Causes:**
1. Workspace directory has restrictive permissions
2. UID mapping failed

**Solutions:**

```bash
# Check workspace permissions
ls -la .

# Try root mode if using rootless Docker
cc-sandbox --root=true claude

# Enable debug output to see what's happening
CC_SANDBOX_DEBUG=1 cc-sandbox claude
```

## Docker Socket Issues

### Docker Socket Access Denied

**Symptom:** `permission denied` when running Docker commands inside the container.

**Solutions:**

1. Ensure you're using the `--docker` flag:
   ```bash
   cc-sandbox -i docker --docker claude
   ```

2. Check socket permissions on host:
   ```bash
   ls -la /var/run/docker.sock
   ```

3. On Linux, ensure your user is in the docker group:
   ```bash
   sudo usermod -aG docker $USER
   # Log out and back in
   ```

### Docker Socket Not Found

**Symptom:** Warning about Docker socket not found.

**Solutions:**

1. Verify Docker is running:
   ```bash
   docker info
   ```

2. Check socket location:
   ```bash
   ls -la /var/run/docker.sock
   ls -la ~/.docker/run/docker.sock      # rootless
   ls -la ~/.orbstack/run/docker.sock    # OrbStack
   ```

3. Set custom socket path:
   ```bash
   CC_SANDBOX_DOCKER_SOCKET=/path/to/docker.sock cc-sandbox --docker claude
   ```

## Image Issues

### Image Not Found

**Symptom:** `Error: No such image: cc-sandbox:base`

**Solutions:**

1. Pull from registry:
   ```bash
   docker pull ghcr.io/luwojtaszek/cc-sandbox:base
   ```

2. Or build locally:
   ```bash
   make images
   ```

3. Check if using correct registry:
   ```bash
   # List local images
   docker images | grep cc-sandbox
   ```

### Image Pull Fails

**Symptom:** Cannot pull images from ghcr.io.

**Solutions:**

1. Check network connectivity:
   ```bash
   curl -I https://ghcr.io
   ```

2. Authenticate with GitHub Container Registry:
   ```bash
   echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
   ```

## Credential Issues

### Claude Authentication Fails

**Symptom:** Claude asks for authentication every time.

**Solutions:**

1. Check if credentials volume exists:
   ```bash
   docker volume ls | grep cc-sandbox
   ```

2. Clear and re-authenticate:
   ```bash
   docker volume rm cc-sandbox-credentials-$(id -u)
   cc-sandbox claude
   ```

3. Check volume contents:
   ```bash
   docker run --rm -v cc-sandbox-credentials-$(id -u):/data alpine ls -la /data
   ```

### Git Identity Not Set

**Symptom:** Git commits fail with "Please tell me who you are".

**Solutions:**

1. Check if gitconfig is mounted:
   ```bash
   CC_SANDBOX_DEBUG=1 cc-sandbox claude
   # Look for .gitconfig mount in output
   ```

2. Override identity:
   ```bash
   cc-sandbox --git-user-name "Your Name" --git-user-email "you@example.com" claude
   ```

3. Set environment variables:
   ```bash
   export CC_SANDBOX_GIT_USER_NAME="Your Name"
   export CC_SANDBOX_GIT_USER_EMAIL="you@example.com"
   ```

## Rootless Docker Issues

### UID Mapping Errors

**Symptom:** Errors about user namespaces or UID mapping.

**Solutions:**

1. Force root mode:
   ```bash
   cc-sandbox --root=true claude
   ```

2. Check if rootless Docker is detected:
   ```bash
   docker info | grep -i rootless
   ```

3. Verify socket location:
   ```bash
   ls -la ~/.docker/run/docker.sock
   ```

### Nested Container Failures

**Symptom:** Docker-in-Docker fails with rootless Docker.

**Solution:** Use host network mode:
```bash
cc-sandbox -i docker --docker --host-network claude
```

## Podman Issues

### Podman Not Detected

**Symptom:** cc-sandbox uses Docker when Podman is preferred.

**Solution:** Explicitly specify runtime:
```bash
cc-sandbox --runtime podman claude
```

### SELinux Denials

**Symptom:** Permission denied errors on SELinux-enabled systems.

**Solutions:**

1. Add `:z` or `:Z` to volume mounts:
   ```bash
   cc-sandbox -m ~/data:/data:z claude
   ```

2. Check SELinux denials:
   ```bash
   ausearch -m avc -ts recent
   ```

## Network Issues

### Cannot Access Host Services

**Symptom:** Cannot connect to services running on host's localhost.

**Solution:** Use host network mode:
```bash
cc-sandbox --host-network claude
```

### DNS Resolution Fails

**Symptom:** Cannot resolve external hostnames.

**Solutions:**

1. Check host DNS:
   ```bash
   host google.com
   ```

2. Use host network:
   ```bash
   cc-sandbox --host-network claude
   ```

## Debug Mode

Enable debug output to see detailed information:

```bash
CC_SANDBOX_DEBUG=1 cc-sandbox claude
```

Debug output includes:
- Runtime detection
- Root mode detection
- Git worktree detection
- Volume mounts
- Container arguments

## Getting Help

If you're still stuck:

1. Search [existing issues](https://github.com/luwojtaszek/cc-sandbox/issues)
2. Open a [new issue](https://github.com/luwojtaszek/cc-sandbox/issues/new) with:
   - Your OS and version
   - Docker/Podman version
   - Debug output (`CC_SANDBOX_DEBUG=1`)
   - Steps to reproduce
