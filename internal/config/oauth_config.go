package config

import (
	"fmt"
	"os"
	"strings"
)

type OAuthConfig struct {
	Enabled        bool     `yaml:"enabled"`
	Provider       string   `yaml:"provider"`
	RedirectURL    string   `yaml:"redirect_url"`
	AllowedDomains []string `yaml:"allowed_domains"`
	AllowedEmails  []string `yaml:"allowed_emails"`
	SessionTimeout string   `yaml:"session_timeout"`
}

type AuthConfig struct {
	OAuth  OAuthConfig     `yaml:"oauth"`
	APIKey string          `yaml:"api_key,omitempty"`
	Basic  BasicAuthConfig `yaml:"basic,omitempty"`
}

func LoadAuthConfig() (*AuthConfig, error) {
	miranteConfig, err := LoadMiranteConfig()
	if err != nil {
		return nil, err
	}
	config := miranteConfig.Auth

	if err := validateAuthConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid auth config: %w", err)
	}

	return &config, nil
}

func validateAuthConfig(config *AuthConfig) error {
	if !config.OAuth.Enabled {
		return nil
	}

	if os.Getenv("OAUTH_CLIENT_ID") == "" {
		return fmt.Errorf("OAUTH_CLIENT_ID environment variable is required when OAuth is enabled")
	}

	if os.Getenv("OAUTH_CLIENT_SECRET") == "" {
		return fmt.Errorf("OAUTH_CLIENT_SECRET environment variable is required when OAuth is enabled")
	}

	if os.Getenv("OAUTH_JWT_SECRET") == "" {
		return fmt.Errorf("OAUTH_JWT_SECRET environment variable is required when OAuth is enabled")
	}

	if config.OAuth.RedirectURL == "" {
		return fmt.Errorf("oauth redirect_url is required when OAuth is enabled")
	}

	if config.OAuth.Provider != "google" && config.OAuth.Provider != "github" {
		return fmt.Errorf("unsupported oauth provider: %s (supported: google, github)", config.OAuth.Provider)
	}

	if len(config.OAuth.AllowedDomains) == 0 && len(config.OAuth.AllowedEmails) == 0 {
		return fmt.Errorf("either allowed_domains or allowed_emails must be specified")
	}

	for _, domain := range config.OAuth.AllowedDomains {
		if !strings.HasPrefix(domain, "@") {
			return fmt.Errorf("domain '%s' must start with '@'", domain)
		}
	}

	return nil
}

func (c *OAuthConfig) IsEmailAllowed(email string) bool {
	for _, allowedEmail := range c.AllowedEmails {
		if strings.EqualFold(email, allowedEmail) {
			return true
		}
	}

	for _, domain := range c.AllowedDomains {
		if strings.HasSuffix(strings.ToLower(email), strings.ToLower(domain)) {
			return true
		}
	}

	return false
}

func CreateSampleAuthConfig() error {
	return fmt.Errorf("auth.yaml is no longer supported; configure auth in %s", GetMiranteConfigPath())
}
