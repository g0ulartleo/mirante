package main

import (
	"context"
	"log"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	alarmrepo "github.com/g0ulartleo/mirante/internal/alarm/repo"
	runtimeclient "github.com/g0ulartleo/mirante/internal/alarm/runtime/client"
	runtimesync "github.com/g0ulartleo/mirante/internal/alarm/runtime/sync"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
	signalrepo "github.com/g0ulartleo/mirante/internal/signal/repo"
	"github.com/g0ulartleo/mirante/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

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

	signalRepo, err := signalrepo.New(config.LoadAppConfigFromEnv())
	if err != nil {
		log.Fatalf("Error initializing signal store: %v", err)
	}
	defer signalRepo.Close()
	signalService := signal.NewService(signalRepo)

	miranteConfig, err := config.LoadMiranteConfig()
	if err != nil {
		log.Fatalf("Error loading mirante config: %v", err)
	}
	runnerClient, err := runtimeclient.NewRouter(miranteConfig.AlarmRuntime)
	if err != nil {
		log.Fatalf("Error initializing alarm runtime router: %v", err)
	}
	defer runnerClient.Close()

	syncer := runtimesync.New(runnerClient, alarmRepo)
	if err := syncer.Sync(context.Background()); err != nil {
		if runtimesync.IsValidationError(err) {
			log.Fatalf("Runtime sync failed: %v", err)
		}
		log.Printf("Warning: runtime sync failed: %v", err)
	}

	go periodicHealthCheck(context.Background(), runnerClient)

	asyncClient := asynq.NewClient(asynq.RedisClientOpt{Addr: config.Env().RedisAddr})
	defer asyncClient.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Env().RedisAddr,
	})
	defer redisClient.Close()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.Env().RedisAddr},
		asynq.Config{
			Concurrency: 10,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("Error processing task %s: %v", task.Type(), err)
			}),
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	worker.RegisterTasks(mux, runnerClient, signalService, alarmService, asyncClient, redisClient)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Error running server: %v", err)
	}
}
