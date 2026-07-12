package alarms

import (
	"fmt"
	alarmsv1 "github.com/g0ulartleo/mirante/packages/go/alarms/v1"
	"google.golang.org/grpc"
	"net"
)

type RuntimeOptions struct {
	Addr    string
	Alarms  []Alarm
	Options []grpc.ServerOption
}

type RuntimeServer struct {
	alarmsv1.UnimplementedAlarmRuntimeServer
	alarms map[string]Alarm
}

func ServeRuntime(opts RuntimeOptions) error {
	addr := opts.Addr
	if addr == "" {
		addr = "127.0.0.1:50051"
	}

	runtimeServer, err := NewRuntimeServer(opts.Alarms)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer(opts.Options...)
	alarmsv1.RegisterAlarmRuntimeServer(grpcServer, runtimeServer)
	return grpcServer.Serve(listener)
}

func NewRuntimeServer(alarms []Alarm) (*RuntimeServer, error) {
	byID := make(map[string]Alarm, len(alarms))
	for _, alarm := range alarms {
		if err := ValidateAlarm(alarm); err != nil {
			return nil, err
		}
		if _, ok := byID[alarm.ID]; ok {
			return nil, fmt.Errorf("duplicate alarm id %q", alarm.ID)
		}
		byID[alarm.ID] = alarm
	}
	return &RuntimeServer{alarms: byID}, nil
}
