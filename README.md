# Traefik Token Authentication Middleware
<img src="./icon.png" alt="Token Authentication Middleware Icon" width="200" height="200">


A Traefik v3 middleware plugin that provides token-based authentication with cookie session support. This plugin is designed for scenarios where you need to protect services with a simple token-based authentication, such as webhook endpoints, preview environments, or internal tools. It allows access to protected routes via a query parameter token and maintains the session using cookies, making it ideal for situations where you want to share access using a URL (e.g., `https://preview.example.com?ta_token=secret-token`) while ensuring the token is not visible in subsequent requests.

> ⚠️ **Important Security Note**: This middleware is intended for simple testing scenarios, preview environments, and quick testing tools. It should not be used as a primary authentication mechanism for production applications or to protect sensitive data. For production environments, please use proper authentication mechanisms like OAuth2, JWT, or other industry-standard protocols.

## Features

- Query parameter token validation
- Session cookie management
- Configurable token list
- URL cleanup (removes token parameter after validation)
- Secure cookie settings (HTTPOnly, Secure, SameSite)

## Configuration

### Static Configuration

To enable the plugin in your Traefik static configuration:

```yaml
experimental:
  plugins:
    tokenauth:
      moduleName: "github.com/Clasyc/tokenauth"
      version: "v0.1.0"
```

### Dynamic Configuration

Example configuration for the middleware:

```yaml
http:
  middlewares:
    token-auth:
      plugin:
        tokenauth:
          tokenParam: "ta_token"           # Query parameter name for the token
          cookieName: "ta_session_token"   # Name of the session cookie
          allowedTokens:                   # List of valid tokens
            - "secret-token-1"
            - "secret-token-2"
```

### Available Options

- `tokenParam`: The query parameter name to look for the token (default: "ta_token")
- `cookieName`: The name of the session cookie (default: "ta_session_token")
- `allowedTokens`: List of tokens that are considered valid

## Usage

### File Configuration

Apply the middleware to your routers in your Traefik dynamic configuration:

```yaml
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      middlewares:
        - token-auth
      service: my-service
```

### Docker Labels

Using Docker labels to configure the middleware:

```yaml
version: '3'

services:
  traefik:
    image: traefik:v3.0
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.docker.exposedbydefault=false"
  
      ## you can enable the plugin by adding the following label
      - "--experimental.plugins.tokenauth.moduleName=github.com/Clasyc/tokenauth"
      - "--experimental.plugins.tokenauth.version=v0.1.0"
    ports:
      - "80:80"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro

  whoami:
    image: traefik/whoami
    labels:
      - "traefik.enable=true"
      # Define the middleware
      - "traefik.http.middlewares.token-auth.plugin.tokenauth.tokenParam=ta_token"
      - "traefik.http.middlewares.token-auth.plugin.tokenauth.cookieName=ta_session_token"
      - "traefik.http.middlewares.token-auth.plugin.tokenauth.allowedTokens=secret-token-1,secret-token-2"
      # Apply middleware to router
      - "traefik.http.routers.whoami.rule=Host(`whoami.localhost`)"
      - "traefik.http.routers.whoami.middlewares=token-auth"
```

Or with Docker run command:

```bash
docker run -d \
  --label "traefik.enable=true" \
  --label "traefik.http.middlewares.token-auth.plugin.tokenauth.tokenParam=ta_token" \
  --label "traefik.http.middlewares.token-auth.plugin.tokenauth.cookieName=ta_session_token" \
  --label "traefik.http.middlewares.token-auth.plugin.tokenauth.allowedTokens=secret-token-1,secret-token-2" \
  --label "traefik.http.routers.my-service.rule=Host(\`example.com\`)" \
  --label "traefik.http.routers.my-service.middlewares=token-auth" \
  my-image
```

### Accessing Protected Services

Once configured, access your service with a token:
- First request: `https://example.com/path?ta_token=secret-token-1`
- Subsequent requests will use the session cookie automatically

## Security Considerations

- Always use HTTPS in production
- Use strong, unique tokens
- Regularly rotate tokens
- Store tokens securely (e.g., using Vault or similar secret management)

## License

MIT License - See LICENSE file for details