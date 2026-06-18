package config

import (
	"fmt"
	"log"
)

type AppConfig struct {
	Driver string      `yaml:"driver"`
	MySQL  MySQLConfig `yaml:"mysql,omitempty"`
}

type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func LoadAppConfigFromEnv() *AppConfig {
	miranteConfig, err := LoadMiranteConfig()
	if err != nil {
		log.Fatalf("failed to load mirante config: %v", err)
	}

	driver := miranteConfig.Storage.Driver
	if driver == "" {
		log.Fatalf("storage.driver is required")
	}

	config := &AppConfig{
		Driver: driver,
	}

	switch driver {
	case "mysql":
		config.MySQL = miranteConfig.Storage.MySQL
	case "sqlite":
		return config
	case "redis":
		return config
	default:
		log.Fatalf("unsupported driver: %s", driver)
	}

	if err := validateConfig(config); err != nil {
		log.Fatalf("invalid app config: %v", err)
	}
	return config
}

func validateConfig(config *AppConfig) error {
	switch config.Driver {
	case "mysql":
		if config.MySQL.Host == "" {
			return fmt.Errorf("mysql host is required")
		}
		if config.MySQL.Port == 0 {
			config.MySQL.Port = 3306
		}
		if config.MySQL.User == "" {
			return fmt.Errorf("mysql user is required")
		}
		if config.MySQL.Password == "" {
			return fmt.Errorf("mysql password is required")
		}
	case "sqlite":
		return nil
	case "redis":
		return nil
	default:
		return fmt.Errorf("unsupported driver: %s", config.Driver)
	}
	return nil
}
