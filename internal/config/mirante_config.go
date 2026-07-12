package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type MiranteConfig struct {
	Storage      StorageConfig      `yaml:"storage"`
	Redis        RedisConfig        `yaml:"redis"`
	HTTP         HTTPConfig         `yaml:"http"`
	AlarmRuntime AlarmRuntimeConfig `yaml:"alarm_runtime"`
	Auth         AuthConfig         `yaml:"auth"`
	SMTP         SMTPConfig         `yaml:"smtp"`
}

type StorageConfig struct {
	Driver string      `yaml:"driver"`
	MySQL  MySQLConfig `yaml:"mysql,omitempty"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
	Port string `yaml:"port"`
}

type AlarmRuntimeConfig struct {
	Timeout  string                   `yaml:"timeout"`
	Runtimes map[string]RuntimeConfig `yaml:"runtimes"`
}

type RuntimeConfig struct {
	Addr    string `yaml:"addr"`
	Version string `yaml:"version,omitempty"`
}

type BasicAuthConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func GetMiranteConfigPath() string {
	if path := os.Getenv("MIRANTE_CONFIG"); path != "" {
		return path
	}
	return "config/mirante.yaml"
}

func LoadMiranteConfig() (*MiranteConfig, error) {
	config := defaultMiranteConfig()
	path := GetMiranteConfigPath()

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read mirante config file: %w", err)
		}

		expanded := os.ExpandEnv(string(data))
		if err := yaml.Unmarshal([]byte(expanded), config); err != nil {
			return nil, fmt.Errorf("failed to parse mirante config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat mirante config file: %w", err)
	}

	applyEnvOverrides(config)
	applyMiranteDefaults(config)
	return config, nil
}

func defaultMiranteConfig() *MiranteConfig {
	return &MiranteConfig{
		Storage: StorageConfig{Driver: "redis"},
		Redis:   RedisConfig{Addr: "127.0.0.1:6379"},
		HTTP:    HTTPConfig{Addr: "127.0.0.1", Port: "40169"},
		AlarmRuntime: AlarmRuntimeConfig{
			Timeout: "30s",
		},
	}
}

func applyEnvOverrides(config *MiranteConfig) {
	if value := os.Getenv("DB_DRIVER"); value != "" {
		config.Storage.Driver = value
	}
	if value := os.Getenv("MYSQL_DB_HOST"); value != "" {
		config.Storage.MySQL.Host = value
	}
	if value := os.Getenv("MYSQL_DB_PORT"); value != "" {
		if port, err := strconv.Atoi(value); err == nil {
			config.Storage.MySQL.Port = port
		}
	}
	if value := os.Getenv("MYSQL_DB_USER"); value != "" {
		config.Storage.MySQL.User = value
	}
	if value := os.Getenv("MYSQL_DB_PASSWORD"); value != "" {
		config.Storage.MySQL.Password = value
	}
	if value := os.Getenv("REDIS_ADDR"); value != "" {
		config.Redis.Addr = value
	}
	if value := os.Getenv("HTTP_ADDR"); value != "" {
		config.HTTP.Addr = value
	}
	if value := os.Getenv("HTTP_PORT"); value != "" {
		config.HTTP.Port = value
	}
	if value := os.Getenv("ALARM_RUNTIME_ADDR"); value != "" {
		config.AlarmRuntime.Runtimes = map[string]RuntimeConfig{"env": {Addr: value}}
	}
	if value := os.Getenv("ALARM_RUNTIME_TIMEOUT"); value != "" {
		config.AlarmRuntime.Timeout = value
	}
	if value := os.Getenv("API_KEY"); value != "" {
		config.Auth.APIKey = value
	}
	if value := os.Getenv("DASHBOARD_BASIC_AUTH_USERNAME"); value != "" {
		config.Auth.Basic.Username = value
	}
	if value := os.Getenv("DASHBOARD_BASIC_AUTH_PASSWORD"); value != "" {
		config.Auth.Basic.Password = value
	}
	if value := os.Getenv("OAUTH_ENABLED"); value != "" {
		config.Auth.OAuth.Enabled = value == "true" || value == "1"
	}
	if value := os.Getenv("OAUTH_PROVIDER"); value != "" {
		config.Auth.OAuth.Provider = value
	}
	if value := os.Getenv("OAUTH_REDIRECT_URL"); value != "" {
		config.Auth.OAuth.RedirectURL = value
	}
	if value := os.Getenv("OAUTH_ALLOWED_DOMAINS"); value != "" {
		config.Auth.OAuth.AllowedDomains = strings.Split(value, ",")
	}
	if value := os.Getenv("OAUTH_ALLOWED_EMAILS"); value != "" {
		config.Auth.OAuth.AllowedEmails = strings.Split(value, ",")
	}
	if value := os.Getenv("OAUTH_SESSION_TIMEOUT"); value != "" {
		config.Auth.OAuth.SessionTimeout = value
	}
	if value := os.Getenv("SMTP_HOST"); value != "" {
		config.SMTP.Host = value
	}
	if value := os.Getenv("SMTP_PORT"); value != "" {
		config.SMTP.Port = value
	}
	if value := os.Getenv("SMTP_USER"); value != "" {
		config.SMTP.User = value
	}
	if value := os.Getenv("SMTP_PASSWORD"); value != "" {
		config.SMTP.Password = value
	}
}

func applyMiranteDefaults(config *MiranteConfig) {
	if config.Storage.Driver == "" {
		config.Storage.Driver = "redis"
	}
	if config.Redis.Addr == "" {
		config.Redis.Addr = "127.0.0.1:6379"
	}
	if config.HTTP.Addr == "" {
		config.HTTP.Addr = "127.0.0.1"
	}
	if config.HTTP.Port == "" {
		config.HTTP.Port = "40169"
	}
	if config.AlarmRuntime.Timeout == "" {
		config.AlarmRuntime.Timeout = "30s"
	}
	if len(config.AlarmRuntime.Runtimes) == 0 {
		config.AlarmRuntime.Runtimes = map[string]RuntimeConfig{"go": {Addr: "127.0.0.1:50051"}}
	}
	if config.Auth.OAuth.Provider == "" {
		config.Auth.OAuth.Provider = "google"
	}
	if config.Auth.OAuth.SessionTimeout == "" {
		config.Auth.OAuth.SessionTimeout = "24h"
	}
}
