package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	alarmrepo "github.com/g0ulartleo/mirante/internal/alarm/repo"
	runtimeclient "github.com/g0ulartleo/mirante/internal/alarm/runtime/client"
	runtimesync "github.com/g0ulartleo/mirante/internal/alarm/runtime/sync"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/worker/tasks"
	"github.com/hibiken/asynq"
)

type AlarmConfigProvider struct {
	alarmService *alarm.AlarmService
}

func (p *AlarmConfigProvider) GetConfigs() ([]*asynq.PeriodicTaskConfig, error) {
	alarms, err := p.alarmService.GetAlarms()
	if err != nil {
		return nil, fmt.Errorf("error getting alarms: %v", err)
	}

	var configs []*asynq.PeriodicTaskConfig
	for _, alarmConfig := range alarms {
		task, err := tasks.NewAlarmCheckTask(alarmConfig.ID)
		if err != nil {
			return nil, fmt.Errorf("error creating alarm check task: %v", err)
		}
		cronspec := alarmConfig.Cron
		if cronspec == "" {
			cronspec = fmt.Sprintf("@every %s", alarmConfig.Interval)
		}
		configs = append(configs, &asynq.PeriodicTaskConfig{
			Cronspec: cronspec,
			Task:     task,
		})
	}

	cleanSignalsTask, err := tasks.NewBackofficeCleanSignalsTask()
	if err != nil {
		return nil, fmt.Errorf("error creating clean signals task: %v", err)
	}
	configs = append(configs, &asynq.PeriodicTaskConfig{
		Cronspec: "@every 24h",
		Task:     cleanSignalsTask,
	})

	return configs, nil
}

func periodicSync(ctx context.Context, syncer *runtimesync.AlarmSyncer) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Println("Running periodic runtime sync")
			if err := syncer.Sync(ctx); err != nil {
				log.Printf("Periodic runtime sync failed: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func periodicHealthCheck(ctx context.Context, router *runtimeclient.Router) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for name, err := range router.Health(ctx) {
				log.Printf("Warning: runtime %q health check failed: %v", name, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	alarmRepo, err := alarmrepo.New()
	if err != nil {
		log.Fatalf("Error initializing alarm store: %v", err)
	}
	defer alarmRepo.Close()

	alarmService := alarm.NewAlarmService(alarmRepo)

	miranteConfig, err := config.LoadMiranteConfig()
	if err != nil {
		log.Fatalf("Error loading mirante config: %v", err)
	}

	router, err := runtimeclient.NewRouter(miranteConfig.AlarmRuntime)
	if err != nil {
		log.Fatalf("Error initializing alarm runtime router: %v", err)
	}
	defer router.Close()

	syncer := runtimesync.New(router, alarmRepo)
	if err := syncer.Sync(context.Background()); err != nil {
		if runtimesync.IsValidationError(err) {
			log.Fatalf("Runtime sync failed: %v", err)
		}
		log.Printf("Warning: runtime sync failed: %v", err)
	}

	go periodicSync(context.Background(), syncer)
	go periodicHealthCheck(context.Background(), router)

	provider := &AlarmConfigProvider{
		alarmService: alarmService,
	}

	mgr, err := asynq.NewPeriodicTaskManager(
		asynq.PeriodicTaskManagerOpts{
			RedisConnOpt:               asynq.RedisClientOpt{Addr: config.Env().RedisAddr},
			PeriodicTaskConfigProvider: provider,
			SyncInterval:               30 * time.Second,
		},
	)
	if err != nil {
		log.Fatalf("Error creating periodic task manager: %v", err)
	}

	if err := mgr.Run(); err != nil {
		log.Fatalf("Error running periodic task manager: %v", err)
	}
}
