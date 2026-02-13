package repo

import (
	"fmt"

	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
)

func New(cfg *config.AppConfig) (signal.SignalRepository, error) {
	switch cfg.Driver {
	case "sqlite":
		return NewSQLiteSignalRepository()
	case "redis":
		return NewRedisSignalRepository()
	case "mysql":
		return NewMySQLSignalRepository(cfg.MySQL)
	case "memory":
		return NewMemorySignalRepository(), nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
	}
}
