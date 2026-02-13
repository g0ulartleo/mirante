# OAuth Authentication Setup Guide

This guide walks you through setting up OAuth authentication for mirante, allowing you to control access using your existing Google or GitHub accounts.

## Quick Setup

1. **Initialize OAuth Configuration**
   ```bash
   make init-oauth
   ```

2. **Configure your OAuth provider** (see detailed steps below)

3. **Update the configuration file** at `config/auth.yaml`

4. **Restart your mirante server**

5. **Authenticate with CLI**
   ```bash
   ./bin/cli auth http://your-domain:40169
   ```

## Detailed Provider Setup

### Google OAuth Setup

1. **Go to Google Cloud Console**
   - Visit [Google Cloud Console](https://console.developers.google.com/)
   - Create a new project or select an existing one

2. **Enable APIs**
   - Go to "APIs & Services" > "Library"
   - Search for "Google+ API" and enable it
   - Also enable "Google OAuth2 API" if available

3. **Create OAuth Credentials**
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth 2.0 Client IDs"
   - Choose "Web application"
   - Add authorized redirect URI: `http://your-domain:40169/auth/callback`
   - Note down the Client ID and Client Secret

4. **Update Configuration**
   ```yaml
   oauth:
     enabled: true
     provider: "google"
     # ... rest of config

5. Add the following environment variables:
   OAUTH_CLIENT_ID=your-oauth-client-id
   OAUTH_CLIENT_SECRET=your-oauth-client-secret
   OAUTH_JWT_SECRET=your-secure-jwt-secret

### GitHub OAuth Setup

1. **Go to GitHub Settings**
   - Visit [GitHub OAuth Apps](https://github.com/settings/applications/new)
   - Or go to Settings > Developer settings > OAuth Apps

2. **Create New OAuth App**
   - Application name: "Mirante Alerts"
   - Homepage URL: `http://your-domain:40169`
   - Authorization callback URL: `http://your-domain:40169/auth/callback`
   - Click "Register application"

3. **Get Credentials**
   - Note down the Client ID
   - Generate a new client secret and note it down

4. **Update Configuration**
   ```yaml
   oauth:
     enabled: true
     provider: "github"
     # ... rest of config

5. Add the following environment variables:
    OAUTH_CLIENT_ID=your-github-client-id
    OAUTH_CLIENT_SECRET=your-github-client-secret
    OAUTH_JWT_SECRET=your-secure-jwt-secret
   ```

## Access Control Configuration

```yaml
oauth:
  # ... other config
  allowed_domains:
    - "@yourcompany.com"  # All company employees
  allowed_emails:
    - "contractor@external.com"  # Specific external contractor
    - "partner@othercorp.com"    # Specific partner
```

## Security Considerations

### JWT Secret Key

Generate a strong, random JWT secret key:

```bash
# Generate a random 32-byte key
openssl rand -base64 32
```

Update the envinronment variable (or .env file):
    OAUTH_JWT_SECRET=your-generated-secret-key

### Session Timeout

Configure appropriate session timeouts:

```yaml
oauth:
  session_timeout: "24h"  # Options: "1h", "8h", "24h", "7d", etc.
```

### HTTPS in Production

For production deployments, always use HTTPS:

```yaml
oauth:
  redirect_url: "https://your-domain.com/auth/callback"
```

Update your OAuth provider settings accordingly.

## CLI Usage

### Initial Authentication

```bash
# Authenticate with OAuth
./bin/cli auth https://your-domain.com

# The command will:
# 1. Open your browser
# 2. Redirect to your OAuth provider
# 3. Ask you to paste the received token
# 4. Save the token locally
```

### Using CLI Commands

After authentication, use CLI commands normally:

```bash
./bin/cli list-alarms
./bin/cli get-alarm my-alarm-id
./bin/cli set-alarm alarm-config.yaml
```

### Token Management

Tokens are stored in `~/.mirante/cli_config.json`. To re-authenticate:

```bash
./bin/cli auth https://your-domain.com
```

## Troubleshooting

### Common Issues

1. **"OAuth is not enabled" error**
   - Check that `enabled: true` in your config
   - Restart the server after config changes

2. **"Invalid redirect URI" error**
   - Ensure the redirect URI in your OAuth provider matches exactly
   - Check for trailing slashes and protocol (http vs https)

3. **"Email not authorized" error**
   - Check your `allowed_domains` and `allowed_emails` configuration
   - Ensure the email format matches (case-sensitive)

4. **"Invalid JWT token" error**
   - Token may have expired, re-authenticate with `./bin/cli auth`
   - Check that `jwt_secret_key` hasn't changed on the server

### Debug Mode

Enable debug logging to troubleshoot issues:

```bash
export LOG_LEVEL=debug

./bin/http-server
```

## Configuration Reference

Complete configuration example:

```yaml
oauth:
  enabled: true
  provider: "google"  # or "github"
  client_id: "your-client-id"
  client_secret: "your-client-secret"
  redirect_url: "https://your-domain.com/auth/callback"
  allowed_domains:
    - "@yourcompany.com"
  allowed_emails:
    - "external@contractor.com"
  jwt_secret_key: "your-secure-random-secret-key"
  session_timeout: "24h"
```
