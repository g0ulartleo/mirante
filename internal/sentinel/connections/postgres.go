package connections

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresConnectionConfig struct {
	URL         string
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	SSLMode     string
	SSLRootCert string
	SSLVerify   bool
}

type PostgresConnection struct {
	DB *sql.DB
}

func NewPostgresConnectionConfig(config map[string]any) (*PostgresConnectionConfig, error) {
	c := PostgresConnectionConfig{}

	if dbURL, ok := config["url"].(string); ok {
		c.URL = dbURL
	}

	if c.URL == "" {
		for _, field := range []string{"host", "port", "user", "password", "database"} {
			if _, ok := config[field]; !ok {
				return nil, fmt.Errorf("missing required field: %s", field)
			}
		}

		c.Host = config["host"].(string)
		c.Port = getIntValue(config["port"])
		c.User = config["user"].(string)
		c.Password = config["password"].(string)
		c.Database = config["database"].(string)
	}

	if sslMode, ok := config["sslmode"].(string); ok {
		c.SSLMode = sslMode
	}

	if sslRootCert, ok := config["sslrootcert"].(string); ok {
		c.SSLRootCert = sslRootCert
	}

	if sslVerify, ok := config["sslverify"].(bool); ok {
		c.SSLVerify = sslVerify
	}

	if c.URL != "" {
		u, err := url.Parse(c.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid postgres connection url: %w", err)
		}
		if u.Scheme != "postgres" && u.Scheme != "postgresql" {
			return nil, fmt.Errorf("invalid postgres connection url scheme: %s", u.Scheme)
		}
		if strings.TrimPrefix(u.Path, "/") == "" {
			return nil, fmt.Errorf("missing required field: database")
		}
	} else {
		if c.Host == "" {
			return nil, fmt.Errorf("missing required field: host")
		}
		if c.Port == 0 {
			c.Port = 5432
		}
		if c.User == "" {
			return nil, fmt.Errorf("missing required field: user")
		}
		if c.Password == "" {
			return nil, fmt.Errorf("missing required field: password")
		}
		if c.Database == "" {
			return nil, fmt.Errorf("missing required field: database")
		}
	}

	return &c, nil
}

func (c *PostgresConnection) Close() error {
	return c.DB.Close()
}

func NewPostgresConnection(config PostgresConnectionConfig) (*PostgresConnection, error) {
	var db *sql.DB
	var err error

	var dsn string
	if config.URL != "" {
		u, parseErr := url.Parse(config.URL)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse postgres connection url: %w", parseErr)
		}

		params := u.Query()
		if config.SSLMode != "" {
			params.Set("sslmode", config.SSLMode)
		}
		if config.SSLRootCert != "" {
			params.Set("sslrootcert", config.SSLRootCert)
		}
		if effectiveSSLMode := params.Get("sslmode"); effectiveSSLMode != "disable" && !config.SSLVerify {
			params.Set("sslrootcert", "")
			params.Set("sslcert", "")
			params.Set("sslkey", "")
		}

		u.RawQuery = params.Encode()
		dsn = u.String()
	} else {
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.Database,
			config.SSLMode,
		)

		if config.SSLMode != "disable" && !config.SSLVerify {
			dsn += " sslrootcert= sslcert= sslkey="
		}
	}

	db, err = sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &PostgresConnection{
		DB: db,
	}, nil
}
