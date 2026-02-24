package main

import (
	"log"
	"net/http"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	alarmrepo "github.com/g0ulartleo/mirante/internal/alarm/repo"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
	signalrepo "github.com/g0ulartleo/mirante/internal/signal/repo"
	"github.com/g0ulartleo/mirante/internal/web/api"
	"github.com/g0ulartleo/mirante/internal/web/dashboard"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	api.RegisterRoutes(e, signalService, alarmService, asyncClient)

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
