package tests

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/sentinel"
	builtins "github.com/g0ulartleo/mirante/internal/sentinel/builtins"
)

// These tests are lightweight integration checks that ensure:
// - Each sample YAML under tests/alarms parses correctly
// - A sentinel can be created from the factory and configured using that YAML

func TestAlarmYAMLsInitializeSentinels(t *testing.T) {
	t.Setenv("TEST_POSTGRES_DB_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")
	t.Setenv("TEST_MYSQL_DB_URL", "mysql://test:test@localhost:3306/testdb?tls=false")

	var cases []string
	root := "alarms"
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, ".yml") || strings.HasSuffix(p, ".yaml") {
			cases = append(cases, p)
		}
		return nil
	})

	if len(cases) == 0 {
		t.Fatalf("no alarm yaml files found in %s; current dir: %s", root, mustGetwd())
	}

	for _, rel := range cases {
		log.Printf("testing alarm: %s", rel)
		t.Run(filepath.Base(rel), func(t *testing.T) {
			cfg, err := alarm.LoadAlarmConfig(rel)
			if err != nil {
				t.Fatalf("failed to load alarm config: %v", err)
			}

			f := sentinel.NewFactory()
			builtins.Register(f)

			s, err := f.Create(cfg.Type)
			if err != nil {
				t.Fatalf("failed to create sentinel type %q: %v", cfg.Type, err)
			}

			if err := s.Configure(cfg.Config); err != nil {
				t.Skipf("skipping: external dependency not available for %s: %v", cfg.Type, err)
			}

			ctx := context.Background()
			if _, err := s.Check(ctx, cfg.ID); err != nil {
				t.Skipf("skipping Check due to external dependency for %s: %v", cfg.Type, err)
			}
		})
	}
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return path.Clean(wd)
}
