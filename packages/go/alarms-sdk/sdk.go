package alarmsdk

import (
	"context"
	"fmt"
	"sort"

	runtimev1 "github.com/g0ulartleo/mirante/packages/go/alarms-sdk/alarmruntime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Signal = runtimev1.RunAlarmResponse

type Alarm struct {
	ID            string
	Name          string
	Description   string
	HowToFix      string
	Path          []string
	Cron          string
	Interval      string
	Notifications *runtimev1.AlarmNotifications
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

func (s *RuntimeServer) Health(ctx context.Context, req *runtimev1.HealthRequest) (*runtimev1.HealthResponse, error) {
	return &runtimev1.HealthResponse{Status: "SERVING"}, nil
}

func (s *RuntimeServer) ListAlarms(ctx context.Context, req *runtimev1.ListAlarmsRequest) (*runtimev1.ListAlarmsResponse, error) {
	ids := make([]string, 0, len(s.alarms))
	for id := range s.alarms {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	items := make([]*runtimev1.Alarm, 0, len(ids))
	for _, id := range ids {
		items = append(items, ToProtoAlarm(s.alarms[id]))
	}
	return &runtimev1.ListAlarmsResponse{Alarms: items}, nil
}

func (s *RuntimeServer) GetAlarm(ctx context.Context, req *runtimev1.GetAlarmRequest) (*runtimev1.GetAlarmResponse, error) {
	alarm, ok := s.alarms[req.GetAlarmId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "alarm %q not found", req.GetAlarmId())
	}
	return &runtimev1.GetAlarmResponse{Alarm: ToProtoAlarm(alarm)}, nil
}

func (s *RuntimeServer) RunAlarm(ctx context.Context, req *runtimev1.RunAlarmRequest) (*runtimev1.RunAlarmResponse, error) {
	if req == nil || req.GetAlarmId() == "" {
		return nil, status.Error(codes.InvalidArgument, "alarm_id is required")
	}

	alarm, ok := s.alarms[req.GetAlarmId()]
	if !ok {
		return errorSignal(runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED, fmt.Sprintf("alarm %q not found", req.GetAlarmId())), nil
	}

	signal, err := alarm.Run(ctx)
	if err != nil {
		return errorSignal(runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_RUNALARM_FAILED, fmt.Sprintf("run failed: %v", err)), nil
	}
	if signal == nil || signal.Status == runtimev1.SignalStatus_SIGNAL_STATUS_UNSPECIFIED {
		return errorSignal(runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_RUNALARM_FAILED, fmt.Sprintf("alarm %q did not return a valid signal", req.GetAlarmId())), nil
	}
	return signal, nil
}

func ToProtoAlarm(alarm Alarm) *runtimev1.Alarm {
	return &runtimev1.Alarm{
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
	return &Signal{Status: runtimev1.SignalStatus_SIGNAL_STATUS_HEALTHY, Message: message}
}

func Unhealthy(message string) *Signal {
	return &Signal{Status: runtimev1.SignalStatus_SIGNAL_STATUS_UNHEALTHY, Message: message}
}

func Warning(message string) *Signal {
	return &Signal{Status: runtimev1.SignalStatus_SIGNAL_STATUS_WARNING, Message: message}
}

func Unknown(message string) *Signal {
	return &Signal{Status: runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN, Message: message}
}

func errorSignal(code runtimev1.RuntimeErrorCode, message string) *Signal {
	return &Signal{
		Status:  runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN,
		Message: message,
		Error: &runtimev1.RuntimeError{
			Code:    code,
			Message: message,
		},
	}
}
