package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	alarmrepo "github.com/g0ulartleo/mirante/internal/alarm/repo"
	runtimeclient "github.com/g0ulartleo/mirante/internal/alarm/runtime/client"
	runtimesync "github.com/g0ulartleo/mirante/internal/alarm/runtime/sync"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
	signalrepo "github.com/g0ulartleo/mirante/internal/signal/repo"
	"github.com/g0ulartleo/mirante/internal/web/api"
	"github.com/g0ulartleo/mirante/internal/web/dashboard"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	miranteConfig, err := config.LoadMiranteConfig()
	if err != nil {
		log.Fatalf("Error loading mirante config: %v", err)
	}

	router, err := runtimeclient.NewRouter(miranteConfig.AlarmRuntime)
	if err != nil {
		log.Fatalf("Error initializing alarm runtime router: %v", err)
	}
	defer router.Close()

	go periodicHealthCheck(context.Background(), router)

	syncer := runtimesync.New(router, alarmRepo)
	if err := syncer.Sync(context.Background()); err != nil {
		if runtimesync.IsValidationError(err) {
			log.Fatalf("Runtime sync failed: %v", err)
		}
		log.Printf("Warning: runtime sync failed: %v", err)
	}

	signalRepo, err := signalrepo.New(config.LoadAppConfigFromEnv())
	if err != nil {
		log.Fatalf("Error initializing signal store: %v", err)
	}
	defer signalRepo.Close()
	signalService := signal.NewService(signalRepo)
	asyncClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr: config.Env().RedisAddr,
	})
	defer asyncClient.Close()

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			log.Printf("HTTP request started method=%s path=%s remote_ip=%s", c.Request().Method, c.Request().URL.Path, c.RealIP())

			err := next(c)

			status := c.Response().Status
			if status == 0 {
				status = http.StatusOK
			}
			if err != nil && status == http.StatusOK {
				status = http.StatusInternalServerError
			}

			log.Printf("HTTP request finished method=%s path=%s status=%d duration=%s error=%v", c.Request().Method, c.Request().URL.Path, status, time.Since(start), err)
			return err
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/static", "static")

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	api.RegisterRoutes(e, signalService, alarmService, asyncClient, syncer)

	dashboardGroup := e.Group("")
	dashboardGroup.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if config.Env().BasicAuthUsername == "" || config.Env().BasicAuthPassword == "" {
			return true, nil
		}
		if username == config.Env().BasicAuthUsername &&
			password == config.Env().BasicAuthPassword {
			return true, nil
		}
		return false, nil
	}))

	dashboardInstance, err := dashboard.NewDashboard(signalService, alarmService, config.Env().RedisAddr)
	if err != nil {
		log.Fatalf("Error initializing dashboard: %v", err)
	}
	defer dashboardInstance.Close()

	dashboardInstance.RegisterRoutes(dashboardGroup)

	e.Logger.Fatal(e.Start(config.Env().HTTPAddr + ":" + config.Env().HTTPPort))
}
