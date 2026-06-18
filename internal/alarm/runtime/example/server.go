package example

import (
	"context"
	"fmt"

	runtimev1 "github.com/g0ulartleo/mirante/proto/alarmruntime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RuntimeServer struct {
	runtimev1.UnimplementedAlarmRuntimeServer
}

func New() *RuntimeServer {
	return &RuntimeServer{}
}

func (s *RuntimeServer) RunAlarm(ctx context.Context, req *runtimev1.RunAlarmRequest) (*runtimev1.RunAlarmResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if req.GetAlarmId() == "" {
		return nil, status.Error(codes.InvalidArgument, "alarm_id is required")
	}

	return &runtimev1.RunAlarmResponse{
		Status:  runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN,
		Message: "no alarms loaded",
		Error: &runtimev1.RuntimeError{
			Code:    runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED,
			Message: fmt.Sprintf("alarm %q not found", req.GetAlarmId()),
		},
	}, nil
}

func (s *RuntimeServer) ListAlarms(ctx context.Context, req *runtimev1.ListAlarmsRequest) (*runtimev1.ListAlarmsResponse, error) {
	return &runtimev1.ListAlarmsResponse{}, nil
}

func (s *RuntimeServer) GetAlarm(ctx context.Context, req *runtimev1.GetAlarmRequest) (*runtimev1.GetAlarmResponse, error) {
	return nil, status.Error(codes.NotFound, "no alarms loaded")
}
