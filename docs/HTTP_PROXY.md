# HTTP Proxy Configuration

## Overview

AI Shell Assistant supports HTTP proxy configuration through standard environment variables. This allows seamless integration with your existing network configuration.

## Configuration

### Environment Variables

The tool automatically reads proxy settings from these environment variables (in priority order):

- `HTTPS_PROXY` - Proxy for HTTPS endpoints
- `https_proxy` - Lowercase version of HTTPS_PROXY
- `HTTP_PROXY` - Proxy for HTTP endpoints
- `http_proxy` - Lowercase version of HTTP_PROXY

### Usage Examples

#### Setting Proxy for All Requests

```bash
# Set proxy for HTTPS requests (most LLM APIs use HTTPS)
export HTTPS_PROXY="http://127.0.0.1:7890"

# Run aiassist
aiassist "check nginx logs"
```

#### Different Proxies for HTTP and HTTPS

```bash
# Set different proxies
export HTTPS_PROXY="http://https-proxy.example.com:8080"
export HTTP_PROXY="http://http-proxy.example.com:8080"

# aiassist will use the appropriate proxy based on the URL protocol
aiassist "analyze system performance"
```

#### Using Lowercase Variables

```bash
# Lowercase variables also work
export https_proxy="http://127.0.0.1:7890"
export http_proxy="http://127.0.0.1:8080"

aiassist "check disk usage"
```

#### No Proxy (Direct Connection)

```bash
# If no proxy environment variables are set, direct connection is used
unset HTTPS_PROXY https_proxy HTTP_PROXY http_proxy

aiassist "show running processes"
```

## How It Works

The tool uses Go's standard `http.ProxyFromEnvironment` function, which:

1. Checks the appropriate environment variable based on the URL protocol
2. For HTTPS URLs: checks `HTTPS_PROXY` or `https_proxy`
3. For HTTP URLs: checks `HTTP_PROXY` or `http_proxy`
4. Returns nil (no proxy) if no environment variable is set

## Timeout Settings

The HTTP client includes comprehensive timeout settings:

- **Total request timeout**: 60 seconds
- **Connection establishment**: 10 seconds
- **TLS handshake**: 10 seconds
- **Response header**: 10 seconds
- **Connection pool**: Up to 10 idle connections, 30-second idle timeout

These settings prevent the tool from hanging indefinitely on network issues.

## Testing

Run the proxy environment variable tests:

```bash
go test -v ./internal/cmd/ -run TestProxyFromEnvironment
```

## Troubleshooting

### Proxy Not Working

1. Check if environment variables are correctly set:
   ```bash
   # Check proxy environment variables
   env | grep -i proxy
   ```

2. Verify proxy server is running:
   ```bash
   curl -x http://127.0.0.1:7890 https://api.openai.com/v1/models
   ```

3. Ensure you're using the correct variable name:
   - For HTTPS APIs (most LLM services): use `HTTPS_PROXY` or `https_proxy`
   - For HTTP APIs: use `HTTP_PROXY` or `http_proxy`

### Timeout Errors

If you see timeout errors, it may indicate:

- Proxy server is not responding
- Network connectivity issues
- API endpoint is unreachable

Try increasing timeout in `internal/llm/openai_compatible.go`:

```go
client := &http.Client{
    Timeout: 120 * time.Second, // Increase from 60s to 120s
    // ...
}
```

## Security Notes

- API keys are sent through the proxy if configured
- Ensure your proxy server is trusted
- TLS verification is enabled by default (recommended)
- Proxy URLs are used directly from environment variables without validation

## NO_PROXY Support

The tool also respects the `NO_PROXY` or `no_proxy` environment variable, which can specify hosts that should bypass the proxy:

```bash
export NO_PROXY="localhost,127.0.0.1,.internal.example.com"
```

This is useful for internal endpoints or development environments.
