help:
	@echo "Available commands:"
	@echo "  go-install-air           Install Air for live reloading"
	@echo "  go-install-templ         Install Templ for template generation"
	@echo "  build                    Build all application components"
	@echo "  build-cli                Build only the CLI binary"
	@echo "  build-cli-all-platforms  Build CLI for all platforms (Linux, macOS, Windows)"
	@echo "  package-cli              Build and package CLI for distribution"
	@echo "  install-cli              Install CLI to /usr/local/bin"
	@echo "  clean-dist               Clean distribution files"
	@echo "  init-oauth               Initialize OAuth configuration"
	@echo "  setup                    Create sample environment configuration"
	@echo "  proto-generate           Generate protobuf Go stubs"

.PHONY: go-install-air
go-install-air:
	go install github.com/air-verse/air@v1.61.7

.PHONY: go-install-templ
go-install-templ:
	go install github.com/a-h/templ/cmd/templ@latest

.PHONY: build
build:
	templ generate
	go build -o ./bin/http-server ./cmd/http-server/server.go
	go build -o ./bin/worker ./cmd/worker-server/main.go
	go build -o ./bin/scheduler ./cmd/scheduler/main.go
	go build -ldflags="-s -w" -o ./bin/mirante ./cmd/cli/main.go

.PHONY: proto-generate
proto-generate:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/alarmruntime/v1/runtime.proto

.PHONY: init-oauth
init-oauth:
	@echo "Initializing OAuth configuration for mirante..."
	@echo ""
	@mkdir -p config
	@if [ ! -f "config/mirante.yaml" ]; then \
		cp examples/config/mirante.yaml config/mirante.yaml; \
		echo "✓ Sample configuration created at config/mirante.yaml"; \
	fi
	@echo "Edit config/mirante.yaml auth.oauth and set enabled: true. auth.yaml is no longer used."
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit config/mirante.yaml to configure allowed users and provider settings"
	@echo ""
	@echo "2. Configure OAuth secrets in your .env file:"
	@echo "   OAUTH_CLIENT_ID=your-oauth-client-id"
	@echo "   OAUTH_CLIENT_SECRET=your-oauth-client-secret"
	@echo "   OAUTH_JWT_SECRET=your-secure-jwt-secret"
	@echo ""
	@echo "3. Configure your OAuth provider (Google/GitHub):"
	@echo "   - Create OAuth application in your provider's console"
	@echo "   - Set redirect URL to: http://your-domain:40169/auth/callback"
	@echo "   - Update OAUTH_CLIENT_ID and OAUTH_CLIENT_SECRET in .env"
	@echo ""
	@echo "4. Configure allowed users in config/mirante.yaml:"
	@echo "   - Update allowed_domains (e.g., ['@yourcompany.com'])"
	@echo "   - Or specify individual allowed_emails"
	@echo ""
	@echo "5. Set auth.oauth.enabled: true in config/mirante.yaml when ready to use OAuth"
	@echo ""
	@echo "For Google OAuth setup:"
	@echo "- Go to: https://console.developers.google.com/"
	@echo "- Create a new project or select existing"
	@echo "- Enable Google+ API"
	@echo "- Create OAuth 2.0 credentials"
	@echo ""
	@echo "For GitHub OAuth setup:"
	@echo "- Go to: https://github.com/settings/applications/new"
	@echo "- Create a new OAuth App"
	@echo "- Set Authorization callback URL"

.PHONY: setup
setup:
	@echo "Creating sample environment configuration..."
	@echo ""
	@if [ -f ".env" ]; then \
		echo "Warning: .env file already exists"; \
		echo "Remove it first if you want to recreate it."; \
		exit 1; \
	fi
	@mkdir -p config/alarms
	@mkdir -p bin
	@if [ ! -f "config/mirante.yaml" ]; then cp examples/config/mirante.yaml config/mirante.yaml; fi
	@printf '# Database Configuration\n# Supported values: "redis", "mysql", "sqlite"\nDB_DRIVER=redis\n\n# MySQL Configuration (only required if DB_DRIVER=mysql)\nMYSQL_DB_HOST=localhost\nMYSQL_DB_PORT=3306\nMYSQL_DB_USER=mirante\nMYSQL_DB_PASSWORD=your-mysql-password\n\n# Redis Configuration\nREDIS_ADDR=127.0.0.1:6379\n\n# HTTP Server Configuration\nHTTP_ADDR=127.0.0.1\nHTTP_PORT=40169\n\n# Email Notifications (SMTP Configuration)\nSMTP_HOST=smtp.gmail.com\nSMTP_PORT=587\nSMTP_USER=your-email@gmail.com\nSMTP_PASSWORD=your-app-password\n\n# Authentication Configuration\n# API key for legacy authentication (use OAuth instead if possible)\nAPI_KEY=your-secure-api-key\n\n# OAuth Configuration (only required if using OAuth)\nOAUTH_CLIENT_ID=your-oauth-client-id\nOAUTH_CLIENT_SECRET=your-oauth-client-secret\nOAUTH_JWT_SECRET=your-secure-jwt-secret\n\n# Basic Auth for Dashboard (optional)\nDASHBOARD_BASIC_AUTH_USERNAME=admin\nDASHBOARD_BASIC_AUTH_PASSWORD=your-dashboard-password\n' > .env
	@echo "✓ Sample environment configuration created at .env"
	@echo "✓ Created necessary directories (config/alarms, bin)"
	@echo ""
	@echo "Next steps:"
	@echo "1. Edit .env file to match your environment:"
	@echo "   - Set DB_DRIVER to your preferred storage backend"
	@echo "   - Configure database connection details"
	@echo "   - Set up SMTP for email notifications (optional)"
	@echo "   - Configure authentication method"
	@echo ""
	@echo "2. For OAuth authentication (recommended):"
	@echo "   make init-oauth"
	@echo ""
	@echo "3. Start the services:"
	@echo "   make build"
	@echo "   docker-compose up -d"
	@echo ""
	@echo "Storage options:"
	@echo "- redis: Fast, in-memory storage (default)"
	@echo "- mysql: Persistent SQL database"
	@echo "- sqlite: Local file-based database"

.PHONY: build-cli-all-platforms
build-cli-all-platforms:
	@echo "Building CLI for all platforms..."
	@mkdir -p ./dist

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/mirante-linux-amd64 ./cmd/cli/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ./dist/mirante-linux-arm64 ./cmd/cli/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/mirante-darwin-amd64 ./cmd/cli/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ./dist/mirante-darwin-arm64 ./cmd/cli/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./dist/mirante-windows-amd64.exe ./cmd/cli/main.go

	@echo "✅ All platform binaries built in ./dist/"

.PHONY: build-cli
build-cli:
	@echo "Building CLI..."
	@mkdir -p ./bin
	go build -ldflags="-s -w" -o ./bin/mirante ./cmd/cli/main.go

.PHONY: package-cli
package-cli: build-cli-all-platforms
	@echo "Creating distribution packages..."
	@cd dist && \
	for file in mirante-*; do \
		if [[ $$file == *.exe ]]; then \
			zip "$${file%.exe}.zip" "$$file" && echo "Created $${file%.exe}.zip"; \
		else \
			tar -czf "$$file.tar.gz" "$$file" && echo "Created $$file.tar.gz"; \
		fi \
	done
	@echo "✅ Distribution packages created in ./dist/"

.PHONY: install-cli
install-cli: build-cli
	@echo "Installing mirante CLI to /usr/local/bin..."
	@if [ -w "/usr/local/bin" ]; then \
		cp ./bin/mirante /usr/local/bin/mirante; \
	else \
		sudo cp ./bin/mirante /usr/local/bin/mirante; \
	fi
	@echo "✅ mirante CLI installed to /usr/local/bin/mirante"

.PHONY: clean-dist
clean-dist:
	@echo "Cleaning distribution files..."
	@rm -rf ./dist
	@echo "✅ Distribution files cleaned"
