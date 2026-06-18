package config

import (
	"fmt"
	"os"
	"sync"
)

type Environment struct {
	DBDriver          string
	MySQLDBPort       string
	MySQLDBHost       string
	MySQLDBUser       string
	MySQLDBPassword   string
	RedisAddr         string
	HTTPPort          string
	HTTPAddr          string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPassword      string
	APIKey            string
	BasicAuthUsername string
	BasicAuthPassword string
	OAuthClientID     string
	OAuthClientSecret string
	OAuthJWTSecret    string
	AlarmRunnerAddr   string
	AlarmRunnerTimeout string
}

var (
	env  *Environment
	once sync.Once
)

func init() {
	InitEnvLoader()
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func firstRuntimeAddr(runtimes map[string]RuntimeConfig) string {
	for _, runtime := range runtimes {
		if runtime.Addr != "" {
			return runtime.Addr
		}
	}
	return "127.0.0.1:50051"
}

func Env() *Environment {
	once.Do(func() {
		miranteConfig, err := LoadMiranteConfig()
		if err != nil {
			miranteConfig = defaultMiranteConfig()
			applyEnvOverrides(miranteConfig)
			applyMiranteDefaults(miranteConfig)
		}

		env = &Environment{
			DBDriver:          miranteConfig.Storage.Driver,
			MySQLDBHost:       miranteConfig.Storage.MySQL.Host,
			MySQLDBPort:       fmt.Sprintf("%d", miranteConfig.Storage.MySQL.Port),
			MySQLDBUser:       miranteConfig.Storage.MySQL.User,
			MySQLDBPassword:   miranteConfig.Storage.MySQL.Password,
			RedisAddr:         miranteConfig.Redis.Addr,
			HTTPPort:          miranteConfig.HTTP.Port,
			HTTPAddr:          miranteConfig.HTTP.Addr,
			SMTPHost:          miranteConfig.SMTP.Host,
			SMTPPort:          miranteConfig.SMTP.Port,
			SMTPUser:          miranteConfig.SMTP.User,
			SMTPPassword:      miranteConfig.SMTP.Password,
			APIKey:            miranteConfig.Auth.APIKey,
			BasicAuthUsername: miranteConfig.Auth.Basic.Username,
			BasicAuthPassword: miranteConfig.Auth.Basic.Password,
			OAuthClientID:     os.Getenv("OAUTH_CLIENT_ID"),
			OAuthClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
			OAuthJWTSecret:    os.Getenv("OAUTH_JWT_SECRET"),
			AlarmRunnerAddr:   firstRuntimeAddr(miranteConfig.AlarmRuntime.Runtimes),
			AlarmRunnerTimeout: miranteConfig.AlarmRuntime.Timeout,
		}
	})

	return env
}
