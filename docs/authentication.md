## authentication

the authentication config for the application lives in the `config/auth.yaml` directory.
mirante currently supports two authentication methods:

### basic api-key

for a simple API-Key setup, set the `API_KEY` environment variable on your server.

```bash
./bin/cli auth-key <your_endpoint> <api_key>
```


### OAuth using a provider (recommended)

OAuth authentication allows you to control access using your existing Google or GitHub accounts, with fine-grained control over who can access the system.

#### setting up OAuth

1. **initialize OAuth
   ```bash
   make init-oauth
   ```
   This creates a sample configuration file at `config/auth.yaml`.

2. **Configure OAuth Provider**

   **For Google OAuth:**
   - Go to [Google Cloud Console](https://console.developers.google.com/)
   - Create a new project or select existing
   - Enable Google+ API
   - Create OAuth 2.0 credentials
   - Set authorized redirect URI to: `http://your-domain:40169/auth/callback`

   **For GitHub OAuth:**
   - Go to [GitHub OAuth Apps](https://github.com/settings/applications/new)
   - Create a new OAuth App
   - Set Authorization callback URL to: `http://your-domain:40169/auth/callback`

3. **Update Configuration**

   **First, configure OAuth secrets in your `.env` file:**
   ```bash
   OAUTH_CLIENT_ID=your-oauth-client-id
   OAUTH_CLIENT_SECRET=your-oauth-client-secret
   OAUTH_JWT_SECRET=your-secure-jwt-secret-key
   ```

   **Then, edit `config/auth.yaml` for non-sensitive settings:**
   ```yaml
   oauth:
     enabled: true
     provider: "google"  # or "github"
     redirect_url: "http://your-domain:40169/auth/callback"
     allowed_domains:
       - "@yourcompany.com"
       - "@contractor.yourcompany.com"
     allowed_emails:
       - "admin@yourcompany.com"
       - "developer@yourcompany.com"
     session_timeout: "24h"
   ```

4. **CLI Authentication**
   ```bash
    mirante auth http://your-domain:40169
   ```
   This will open your browser for authentication and save the token locally.
