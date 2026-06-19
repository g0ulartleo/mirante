package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/alarm/runtime/sync"
	"github.com/g0ulartleo/mirante/internal/auth"
	"github.com/g0ulartleo/mirante/internal/config"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/g0ulartleo/mirante/internal/web/dashboard"
	"github.com/g0ulartleo/mirante/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
)

func APIKeyAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey != config.Env().APIKey {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}
			return next(c)
		}
	}
}

func RegisterRoutes(e *echo.Echo, signalService *signal.Service, alarmService *alarm.AlarmService, asyncClient *asynq.Client, syncer *sync.AlarmSyncer) {
	authConfig, err := config.LoadAuthConfig()
	if err != nil {
		log.Printf("Error loading auth config, using environment API key: %v", err)
		authConfig = &config.AuthConfig{
			OAuth:  config.OAuthConfig{Enabled: false},
			APIKey: config.Env().APIKey,
		}
	}

	if authConfig.OAuth.Enabled {
		oauthService, err := auth.NewOAuthService(authConfig)
		if err != nil {
			log.Printf("Error creating OAuth service: %v", err)
		} else {
			oauthHandlers := auth.NewOAuthHandlers(oauthService)

			e.GET("/auth/login", oauthHandlers.LoginHandler, auth.LoginRateLimitMiddleware(5))
			e.GET("/auth/callback", oauthHandlers.CallbackHandler, auth.AuthRateLimitMiddleware(10))
			e.POST("/auth/logout", oauthHandlers.LogoutHandler, auth.AuthRateLimitMiddleware(10))
			e.GET("/auth/status", oauthHandlers.StatusHandler, auth.AuthRateLimitMiddleware(10))
		}
	}

	api := e.Group("/api")

	if authConfig.OAuth.Enabled || authConfig.APIKey != "" {
		api.Use(auth.AuthenticationMiddleware())
	} else {
		log.Println("Warning: No authentication method configured")
	}

	api.Use(auth.AuthRateLimitMiddleware(45))

	api.GET("/alarms/signals", func(c echo.Context) error {
		alarmSignals, err := dashboard.GetAlarmSignals(signalService, alarmService)
		if err != nil {
			log.Printf("Error fetching config signals: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, alarmSignals)
	})

	jsonWSBroker := NewJSONWebSocketBroker(config.Env().RedisAddr)
	api.GET("/alarms/ws", HandleJSONWebSocket(jsonWSBroker, signalService, alarmService))

	api.GET("/alarms/:alarm_id/signals", func(c echo.Context) error {
		alarmID := c.Param("alarm_id")
		if v := c.QueryParam("since"); v != "" {
			since, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid since timestamp")
			}
			alarmSignals, err := signalService.GetAlarmSignalsSince(alarmID, since)
			if err != nil {
				log.Printf("Error fetching config signals: %v", err)
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			return c.JSON(http.StatusOK, alarmSignals)
		}

		limit := 10
		if v := c.QueryParam("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				if n > 100 {
					n = 100
				}
				limit = n
			}
		}
		alarmSignals, err := signalService.GetAlarmLatestSignals(alarmID, limit)
		if err != nil {
			log.Printf("Error fetching config signals: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, alarmSignals)
	})

	api.GET("/alarms", func(c echo.Context) error {
		alarms, err := alarmService.GetAlarms()
		if err != nil {
			log.Printf("Error fetching config signals: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		maskedAlarms := make([]*alarm.Alarm, 0, len(alarms))
		for _, a := range alarms {
			maskedAlarms = append(maskedAlarms, MaskSensitiveData(a))
		}
		return c.JSON(http.StatusOK, maskedAlarms)
	})

	api.GET("/alarms/:alarm_id", func(c echo.Context) error {
		alarmID := c.Param("alarm_id")
		alarm, err := alarmService.GetAlarm(alarmID)
		if err != nil {
			log.Printf("Error fetching config signals: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		maskedAlarm := MaskSensitiveData(alarm)
		return c.JSON(http.StatusOK, maskedAlarm)
	})

	api.POST("/alarms/sync", func(c echo.Context) error {
		log.Printf("Manual alarm sync requested remote_ip=%s", c.RealIP())
		if err := syncer.Sync(c.Request().Context()); err != nil {
			log.Printf("Error syncing alarms: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		log.Printf("Manual alarm sync completed")
		return c.JSON(http.StatusOK, map[string]string{"message": "Sync completed"})
	})

	api.POST("/alarms/:alarm_id/check", func(c echo.Context) error {
		alarmID := c.Param("alarm_id")
		log.Printf("Manual alarm check requested alarm_id=%s remote_ip=%s", alarmID, c.RealIP())
		task, err := tasks.NewAlarmCheckTask(alarmID)
		if err != nil {
			log.Printf("Error creating check alarm task: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if _, err := asyncClient.Enqueue(task); err != nil {
			log.Printf("Error enqueueing task: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		log.Printf("Manual alarm check enqueued alarm_id=%s", alarmID)
		return c.JSON(http.StatusOK, map[string]string{"message": "Task enqueued"})
	}, auth.AuthRateLimitMiddleware(10))
}
