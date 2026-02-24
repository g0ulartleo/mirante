package api

import (
	"log"
	"net/http"

	"github.com/g0ulartleo/mirante/internal/alarm"
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

func RegisterRoutes(e *echo.Echo, signalService *signal.Service, alarmService *alarm.AlarmService, asyncClient *asynq.Client) {
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

	api.GET("/alarms/:alarm_id/signals", func(c echo.Context) error {
		alarmID := c.Param("alarm_id")
		alarmSignals, err := signalService.GetAlarmLatestSignals(alarmID, 10)
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
		return c.JSON(http.StatusOK, alarms)
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

	api.DELETE("/alarms/:alarm_id", func(c echo.Context) error {
		alarmID := c.Param("alarm_id")
		if err := alarmService.DeleteAlarm(alarmID); err != nil {
			log.Printf("Error deleting alarm: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Alarm deleted"})
	})

	api.POST("/alarms", func(c echo.Context) error {
		alarm := new(alarm.Alarm)
		if err := c.Bind(alarm); err != nil {
			log.Printf("Error binding alarm: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err := alarmService.SetAlarm(alarm); err != nil {
			log.Printf("Error setting alarm: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, alarm)
	})

	api.POST("/alarms/:alarm_id/check", func(c echo.Context) error {
		alarmID := c.Param("alarm_id")
		log.Printf("Manual sentinel check requested alarm_id=%s remote_ip=%s", alarmID, c.RealIP())
		task, err := tasks.NewAlarmCheckTask(alarmID)
		if err != nil {
			log.Printf("Error creating check alarm task: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if _, err := asyncClient.Enqueue(task); err != nil {
			log.Printf("Error enqueueing task: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		log.Printf("Manual sentinel check enqueued alarm_id=%s", alarmID)
		return c.JSON(http.StatusOK, map[string]string{"message": "Task enqueued"})
	}, auth.AuthRateLimitMiddleware(10))
}
