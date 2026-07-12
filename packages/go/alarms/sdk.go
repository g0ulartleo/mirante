package alarms

import (
	"context"
	"fmt"
	"sort"

	alarmsv1 "github.com/g0ulartleo/mirante/packages/go/alarms/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Signal = alarmsv1.RunAlarmResponse

type Alarm struct {
	ID            string
	Name          string
	Description   string
	HowToFix      string
	Path          []string
	Cron          string
	Interval      string
	Notifications *alarmsv1.AlarmNotifications
	Run           func(context.Context) (*Signal, error)
}

func ValidateAlarm(alarm Alarm) error {
	if alarm.ID == "" {
		return fmt.Errorf("alarm id is required")
	}
	if alarm.Description == "" {
		return fmt.Errorf("alarm %q description is required", alarm.ID)
	}
	if alarm.Cron == "" && alarm.Interval == "" {
		return fmt.Errorf("alarm %q cron or interval is required", alarm.ID)
	}
	if alarm.Cron != "" && alarm.Interval != "" {
		return fmt.Errorf("alarm %q cron and interval cannot both be set", alarm.ID)
	}
	if alarm.Run == nil {
		return fmt.Errorf("alarm %q run function is required", alarm.ID)
	}
	return nil
}

func (s *RuntimeServer) Health(ctx context.Context, req *alarmsv1.HealthRequest) (*alarmsv1.HealthResponse, error) {
	return &alarmsv1.HealthResponse{Status: "SERVING"}, nil
}

func (s *RuntimeServer) ListAlarms(ctx context.Context, req *alarmsv1.ListAlarmsRequest) (*alarmsv1.ListAlarmsResponse, error) {
	ids := make([]string, 0, len(s.alarms))
	for id := range s.alarms {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	items := make([]*alarmsv1.Alarm, 0, len(ids))
	for _, id := range ids {
		items = append(items, ToProtoAlarm(s.alarms[id]))
	}
	return &alarmsv1.ListAlarmsResponse{Alarms: items}, nil
}

func (s *RuntimeServer) GetAlarm(ctx context.Context, req *alarmsv1.GetAlarmRequest) (*alarmsv1.GetAlarmResponse, error) {
	alarm, ok := s.alarms[req.GetAlarmId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "alarm %q not found", req.GetAlarmId())
	}
	return &alarmsv1.GetAlarmResponse{Alarm: ToProtoAlarm(alarm)}, nil
}

func (s *RuntimeServer) RunAlarm(ctx context.Context, req *alarmsv1.RunAlarmRequest) (*alarmsv1.RunAlarmResponse, error) {
	if req == nil || req.GetAlarmId() == "" {
		return nil, status.Error(codes.InvalidArgument, "alarm_id is required")
	}

	alarm, ok := s.alarms[req.GetAlarmId()]
	if !ok {
		return errorSignal(alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED, fmt.Sprintf("alarm %q not found", req.GetAlarmId())), nil
	}

	signal, err := alarm.Run(ctx)
	if err != nil {
		return errorSignal(alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_RUNALARM_FAILED, fmt.Sprintf("run failed: %v", err)), nil
	}
	if signal == nil || signal.Status == alarmsv1.SignalStatus_SIGNAL_STATUS_UNSPECIFIED {
		return errorSignal(alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_RUNALARM_FAILED, fmt.Sprintf("alarm %q did not return a valid signal", req.GetAlarmId())), nil
	}
	return signal, nil
}

func ToProtoAlarm(alarm Alarm) *alarmsv1.Alarm {
	return &alarmsv1.Alarm{
		Id:            alarm.ID,
		Name:          alarm.Name,
		Description:   alarm.Description,
		HowToFix:      alarm.HowToFix,
		Path:          alarm.Path,
		Cron:          alarm.Cron,
		Interval:      alarm.Interval,
		Notifications: alarm.Notifications,
	}
}

func Healthy(message string) *Signal {
	return &Signal{Status: alarmsv1.SignalStatus_SIGNAL_STATUS_HEALTHY, Message: message}
}

func Unhealthy(message string) *Signal {
	return &Signal{Status: alarmsv1.SignalStatus_SIGNAL_STATUS_UNHEALTHY, Message: message}
}

func Warning(message string) *Signal {
	return &Signal{Status: alarmsv1.SignalStatus_SIGNAL_STATUS_WARNING, Message: message}
}

func Unknown(message string) *Signal {
	return &Signal{Status: alarmsv1.SignalStatus_SIGNAL_STATUS_UNKNOWN, Message: message}
}

func errorSignal(code alarmsv1.RuntimeErrorCode, message string) *Signal {
	return &Signal{
		Status:  alarmsv1.SignalStatus_SIGNAL_STATUS_UNKNOWN,
		Message: message,
		Error: &alarmsv1.RuntimeError{
			Code:    code,
			Message: message,
		},
	}
}
