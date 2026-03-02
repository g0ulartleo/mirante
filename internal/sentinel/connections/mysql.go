package connections

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh"
)

type MySQLConnectionConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Params   map[string]string
	Tunnel   TunnelConfig
}

type MySQLConnection struct {
	DB        *sql.DB
	sshClient *ssh.Client
}

func getIntValue(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		log.Fatalf("warning: expected value is not an int or float: %v", value)
		return 0
	}
}

func NewMySQLConnectionConfig(config map[string]any) (*MySQLConnectionConfig, error) {
	var c MySQLConnectionConfig

	if dbURL, ok := config["url"].(string); ok {
		c.URL = dbURL
	}

	if c.URL != "" {
		parsedConfig, err := newMySQLConfigFromURL(c.URL)
		if err != nil {
			return nil, err
		}
		c = *parsedConfig
	} else {
		for _, field := range []string{"host", "port", "user", "password", "database"} {
			if _, ok := config[field]; !ok {
				return nil, fmt.Errorf("missing required field: %s", field)
			}
		}
		c = MySQLConnectionConfig{
			Host:     config["host"].(string),
			Port:     getIntValue(config["port"]),
			User:     config["user"].(string),
			Password: config["password"].(string),
			Database: config["database"].(string),
		}
	}

	if tunnelConfig, ok := config["tunnel"].(map[string]any); ok {
		tunnel, err := NewTunnelConfig(tunnelConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create tunnel config: %v", err)
		}
		c.Tunnel = *tunnel
	}
	if c.Host == "" {
		return nil, fmt.Errorf("missing required field: host")
	}
	if c.Port == 0 {
		c.Port = 3306
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
	return &c, nil
}

func (c *MySQLConnection) Close() error {
	var dbErr, sshErr error
	if c.DB != nil {
		dbErr = c.DB.Close()
	}
	if c.sshClient != nil {
		sshErr = c.sshClient.Close()
	}
	if dbErr != nil {
		return dbErr
	}
	return sshErr
}

func NewMySQLConnection(config MySQLConnectionConfig) (*MySQLConnection, error) {
	var db *sql.DB
	var sshClient *ssh.Client
	var err error

	baseDSNConfig := mysql.NewConfig()
	baseDSNConfig.User = config.User
	baseDSNConfig.Passwd = config.Password
	baseDSNConfig.DBName = config.Database
	baseDSNConfig.Params = config.Params

	if config.Tunnel.Host != "" {
		sshClient, err = NewSSHClient(config.Tunnel)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH client: %v", err)
		}

		dialName := "mysql+ssh"
		mysql.RegisterDialContext(dialName, func(ctx context.Context, addr string) (net.Conn, error) {
			return sshClient.Dial("tcp", addr)
		})

		baseDSNConfig.Net = dialName
		baseDSNConfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
		dsn := baseDSNConfig.FormatDSN()

		db, err = sql.Open("mysql", dsn)
		if err != nil {
			sshClient.Close()
			return nil, fmt.Errorf("failed to connect to database via SSH tunnel: %v", err)
		}
	} else {
		baseDSNConfig.Net = "tcp"
		baseDSNConfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
		dsn := baseDSNConfig.FormatDSN()

		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %v", err)
		}
	}

	return &MySQLConnection{
		DB:        db,
		sshClient: sshClient,
	}, nil
}

func newMySQLConfigFromURL(rawURL string) (*MySQLConnectionConfig, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid mysql connection url: %w", err)
	}
	if u.Scheme != "mysql" {
		return nil, fmt.Errorf("invalid mysql connection url scheme: %s", u.Scheme)
	}

	port := 3306
	if u.Port() != "" {
		parsedPort, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, fmt.Errorf("invalid mysql connection url port: %w", err)
		}
		port = parsedPort
	}

	password, _ := u.User.Password()
	database := strings.TrimPrefix(u.Path, "/")
	params := make(map[string]string, len(u.Query()))
	for k, values := range u.Query() {
		if len(values) == 0 {
			continue
		}
		params[k] = values[len(values)-1]
	}

	cfg := &MySQLConnectionConfig{
		URL:      rawURL,
		Host:     u.Hostname(),
		Port:     port,
		User:     u.User.Username(),
		Password: password,
		Database: database,
		Params:   params,
	}

	if cfg.Host == "" {
		return nil, fmt.Errorf("missing required field: host")
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("missing required field: user")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("missing required field: password")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("missing required field: database")
	}

	return cfg, nil
}
