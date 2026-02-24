package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/g0ulartleo/mirante/internal/sentinel"
	"github.com/g0ulartleo/mirante/internal/signal"
	runtimev1 "github.com/g0ulartleo/mirante/proto/sentinelruntime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	runtimev1.UnimplementedSentinelRuntimeServer
	factory *sentinel.SentinelFactory
}

func New(factory *sentinel.SentinelFactory) *Server {
	return &Server{factory: factory}
}

func (s *Server) Check(ctx context.Context, req *runtimev1.CheckRequest) (*runtimev1.CheckResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	if req.GetAlarmId() == "" {
		return nil, status.Error(codes.InvalidArgument, "alarm_id is required")
	}
	if req.GetSentinelType() == "" {
		return nil, status.Error(codes.InvalidArgument, "sentinel_type is required")
	}

	checker, err := s.factory.Create(req.GetSentinelType())
	if err != nil {
		return errorResponse(
			runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_UNSUPPORTED_TYPE,
			fmt.Sprintf("unsupported sentinel type %q", req.GetSentinelType()),
		), nil
	}

	var config map[string]any
	if len(req.GetConfigJson()) > 0 {
		if err := json.Unmarshal(req.GetConfigJson(), &config); err != nil {
			return errorResponse(
				runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_INVALID_CONFIG,
				fmt.Sprintf("invalid config_json: %v", err),
			), nil
		}
	}

	if err := checker.Configure(config); err != nil {
		return errorResponse(
			runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_INVALID_CONFIG,
			fmt.Sprintf("failed to configure sentinel: %v", err),
		), nil
	}

	sig, err := checker.Check(ctx, req.GetAlarmId())
	if err != nil {
		return errorResponse(
			runtimev1.RuntimeErrorCode_RUNTIME_ERROR_CODE_CHECK_FAILED,
			fmt.Sprintf("failed to check sentinel: %v", err),
		), nil
	}

	return &runtimev1.CheckResponse{
		Status:  toProtoStatus(sig.Status),
		Message: sig.Message,
	}, nil
}

func errorResponse(code runtimev1.RuntimeErrorCode, message string) *runtimev1.CheckResponse {
	return &runtimev1.CheckResponse{
		Status:  runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN,
		Message: message,
		Error: &runtimev1.RuntimeError{
			Code:    code,
			Message: message,
		},
	}
}

func toProtoStatus(status signal.Status) runtimev1.SignalStatus {
	switch status {
	case signal.StatusHealthy:
		return runtimev1.SignalStatus_SIGNAL_STATUS_HEALTHY
	case signal.StatusUnhealthy:
		return runtimev1.SignalStatus_SIGNAL_STATUS_UNHEALTHY
	default:
		return runtimev1.SignalStatus_SIGNAL_STATUS_UNKNOWN
	}
}
