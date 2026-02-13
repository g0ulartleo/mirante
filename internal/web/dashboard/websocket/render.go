package websocket

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/a-h/templ"

	"github.com/g0ulartleo/mirante/internal/alarm"
	"github.com/g0ulartleo/mirante/internal/web/dashboard/templates"
)

func RenderComponent(currentPath string, message []byte) (templ.Component, error) {
	var alarmSignals []alarm.AlarmSignals
	if err := json.Unmarshal(message, &alarmSignals); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}
	if currentPath == "/" {
		component := templates.Alarms(alarmSignals)
		return component, nil
	}
	segments := strings.Split(currentPath, "/")
	level := len(segments)
	unescapedPath, err := url.PathUnescape(currentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape path: %v", err)
	}
	component := templates.Base(templates.Treemap(alarmSignals, level, unescapedPath))
	return component, nil
}

func getPathFromQueryString(queryString string) string {
	values, err := url.ParseQuery(queryString)
	if err != nil {
		return ""
	}
	path := values.Get("path")
	if path == "" {
		return "/"
	}
	return path
}
