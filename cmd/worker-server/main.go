package main

import (
	"context"
	"log"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	alarmrepo "github.com/g0ulartleo/mirante/internal/alarm/repo"
	"github.com/g0ulartleo/mirante/internal/config"
	runtimeclient "github.com/g0ulartleo/mirante/internal/sentinel/runtime/client"
	"github.com/g0ulartleo/mirante/internal/signal"
	signalrepo "github.com/g0ulartleo/mirante/internal/signal/repo"
	"github.com/g0ulartleo/mirante/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func main() {
	alarmRepo, err := alarmrepo.New()
	if err != nil {
		log.Fatalf("Error initializing alarm store: %v", err)
	}
	defer alarmRepo.Close()
	alarmService := alarm.NewAlarmService(alarmRepo)
	err = alarm.InitAlarms(alarmRepo)
	if err != nil {
		log.Fatalf("Error initializing alarm configs: %v", err)
	}

	signalRepo, err := signalrepo.New(config.LoadAppConfigFromEnv())
	if err != nil {
		log.Fatalf("Error initializing signal store: %v", err)
	}
	defer signalRepo.Close()
	signalService := signal.NewService(signalRepo)

	runnerTimeout, err := time.ParseDuration(config.Env().SentinelRunnerTimeout)
	if err != nil {
		log.Fatalf("Invalid SENTINEL_RUNNER_TIMEOUT %q: %v", config.Env().SentinelRunnerTimeout, err)
	}
	runnerClient, err := runtimeclient.New(config.Env().SentinelRunnerAddr, runnerTimeout)
	if err != nil {
		log.Fatalf("Error initializing sentinel runner client: %v", err)
	}
	defer runnerClient.Close()

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
