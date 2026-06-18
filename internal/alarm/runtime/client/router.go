package client

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
)

type Router struct {
	clients []namedClient
	byName  map[string]namedClient
}

type namedClient struct {
	name   string
	addr   string
	client *Client
}

func NewRouter(runtimeConfig config.AlarmRuntimeConfig) (*Router, error) {
	timeout, err := time.ParseDuration(runtimeConfig.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid alarm runtime timeout %q: %w", runtimeConfig.Timeout, err)
	}

	names := make([]string, 0, len(runtimeConfig.Runtimes))
	for name := range runtimeConfig.Runtimes {
		names = append(names, name)
	}
	sort.Strings(names)

	clients := make([]namedClient, 0, len(names))
	byName := make(map[string]namedClient, len(names))
	for _, name := range names {
		runtime := runtimeConfig.Runtimes[name]
		if strings.TrimSpace(runtime.Addr) == "" {
			return nil, fmt.Errorf("alarm runtime %q addr is required", name)
		}

		client, err := New(runtime.Addr, timeout)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize alarm runtime %q at %s: %w", name, runtime.Addr, err)
		}
		nc := namedClient{name: name, addr: runtime.Addr, client: client}
		clients = append(clients, nc)
		byName[name] = nc
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("at least one alarm runtime is required")
	}

	return &Router{clients: clients, byName: byName}, nil
}

func (r *Router) Close() error {
	var errs []string
	for _, client := range r.clients {
		if err := client.client.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", client.name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close alarm runtime clients: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (r *Router) ListAlarms(ctx context.Context) ([]*alarm.Alarm, error) {
	byRuntime, err := r.ListAlarmsByRuntime(ctx)
	var all []*alarm.Alarm
	for _, alarms := range byRuntime {
		all = append(all, alarms...)
	}
	return all, err
}

func (r *Router) ListAlarmsByRuntime(ctx context.Context) (map[string][]*alarm.Alarm, error) {
	byRuntime := map[string][]*alarm.Alarm{}
	var errs []string
	for _, nc := range r.clients {
		alarms, err := nc.client.ListAlarms(ctx)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s(%s): %v", nc.name, nc.addr, err))
			continue
		}
		for _, a := range alarms {
			a.Runtime = nc.name
		}
		byRuntime[nc.name] = alarms
	}
	if len(errs) > 0 {
		return byRuntime, fmt.Errorf("failed to list alarms from runtimes: %s", strings.Join(errs, "; "))
	}
	return byRuntime, nil
}

func (r *Router) RunAlarm(ctx context.Context, runtimeName string, alarmID string) (signal.Signal, error) {
	nc, ok := r.byName[runtimeName]
	if !ok {
		return signal.Signal{}, fmt.Errorf("alarm runtime %q not found", runtimeName)
	}

	sig, err := nc.client.RunAlarm(ctx, alarmID)
	if err == nil {
		return sig, nil
	}

	runtimeErr, ok := err.(*RuntimeError)
	if ok {
		return sig, runtimeErr
	}

	return signal.Signal{}, err
}

func (r *Router) Health(ctx context.Context) map[string]error {
	results := map[string]error{}
	for _, nc := range r.clients {
		if err := nc.client.Health(ctx); err != nil {
			results[nc.name] = err
		}
	}
	return results
}
