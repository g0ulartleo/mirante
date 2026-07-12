package example

import (
	"context"
	"fmt"

	alarmsv1 "github.com/g0ulartleo/mirante/proto/alarms/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RuntimeServer struct {
	alarmsv1.UnimplementedAlarmRuntimeServer
}

func New() *RuntimeServer {
	return &RuntimeServer{}
}

func (s *RuntimeServer) RunAlarm(ctx context.Context, req *alarmsv1.RunAlarmRequest) (*alarmsv1.RunAlarmResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if req.GetAlarmId() == "" {
		return nil, status.Error(codes.InvalidArgument, "alarm_id is required")
	}

	return &alarmsv1.RunAlarmResponse{
		Status:  alarmsv1.SignalStatus_SIGNAL_STATUS_UNKNOWN,
		Message: "no alarms loaded",
		Error: &alarmsv1.RuntimeError{
			Code:    alarmsv1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED,
			Message: fmt.Sprintf("alarm %q not found", req.GetAlarmId()),
		},
	}, nil
}

func (s *RuntimeServer) ListAlarms(ctx context.Context, req *alarmsv1.ListAlarmsRequest) (*alarmsv1.ListAlarmsResponse, error) {
	return &alarmsv1.ListAlarmsResponse{}, nil
}

func (s *RuntimeServer) GetAlarm(ctx context.Context, req *alarmsv1.GetAlarmRequest) (*alarmsv1.GetAlarmResponse, error) {
	return nil, status.Error(codes.NotFound, "no alarms loaded")
}
