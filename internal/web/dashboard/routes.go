package dashboard

import (
	"log"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/signal"
	"github.com/g0ulartleo/mirante/internal/web/dashboard/templates"
	"github.com/g0ulartleo/mirante/internal/web/dashboard/websocket"
	"github.com/labstack/echo/v4"
)

type Dashboard struct {
	wsBroker      *websocket.WebSocketBroker
	signalService *signal.Service
	alarmService  *alarm.AlarmService
}

func NewDashboard(signalService *signal.Service, alarmService *alarm.AlarmService, redisAddr string) (*Dashboard, error) {
	wsBroker, err := websocket.NewWebSocketBroker(redisAddr)
	if err != nil {
		return nil, err
	}
	go wsBroker.Run()
	return &Dashboard{
		wsBroker:      wsBroker,
		signalService: signalService,
		alarmService:  alarmService,
	}, nil
}

func (d *Dashboard) RegisterRoutes(dashboard *echo.Group) {
	dashboard.GET("/", func(c echo.Context) error {
		alarmSignals, err := GetAlarmSignals(d.signalService, d.alarmService)
		if err != nil {
			log.Printf("Error fetching config signals: %v", err)
			return RenderError(c, http.StatusInternalServerError, err)
		}
		return Render(c, http.StatusOK, templates.Alarms(alarmSignals))
	})

	dashboard.GET("/ws", websocket.HandleWebSocket(d.wsBroker))

	dashboard.GET("/*", func(c echo.Context) error {
		pathParam := c.Param("*")
		var level int
		var baseURL string
		if pathParam == "" {
			level = 0
			baseURL = "/"
		} else {
			segments := strings.Split(pathParam, "/")
			level = len(segments)
			baseURL = "/" + pathParam
		}
		alarmSignals, err := GetAlarmSignals(d.signalService, d.alarmService)
		if err != nil {
			log.Printf("Error fetching config signals: %v", err)
			return RenderError(c, http.StatusInternalServerError, err)
		}

		return Render(c, http.StatusOK, templates.Treemap(alarmSignals, level, baseURL))
	})
}

func (d *Dashboard) Close() error {
	return d.wsBroker.Close()
}

func Render(ctx echo.Context, statusCode int, template templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)
	if ctx.Path() != "/" {
		template = templates.Base(template)
	}
	if err := template.Render(ctx.Request().Context(), buf); err != nil {
		return RenderError(ctx, http.StatusInternalServerError, err)
	}
	return ctx.HTML(statusCode, buf.String())
}

func RenderError(ctx echo.Context, statusCode int, err error) error {
	return ctx.HTML(statusCode, err.Error())
}
