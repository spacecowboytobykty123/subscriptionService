package grpcapp

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	subgrpc "subscriptionMService/internal/grpc/subscription"
	"subscriptionMService/internal/jsonlog"
)

type App struct {
	Log        *jsonlog.Logger
	GRPCServer *grpc.Server
	Port       int
}

func New(log *jsonlog.Logger, port int, subService subgrpc.Subscription) *App {
	gRPCServer := grpc.NewServer()

	subgrpc.Register(gRPCServer, subService)

	return &App{
		Log:        log,
		GRPCServer: gRPCServer,
		Port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", "grpcapp.Run", err)
	}
	a.Log.PrintInfo("Running GRPC server", nil)

	if err := a.GRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s:%d", "grpcapp.Run", err)
	}
	return nil
}

func (a *App) Stop() {
	a.GRPCServer.GracefulStop()
}
