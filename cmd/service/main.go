package main

import (
	"context"
	"fmt"
	prometheusclient "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"matchmaking/internal/api"
	"matchmaking/internal/app"
	"matchmaking/internal/matchmaking"
	"matchmaking/internal/metrics"
	"matchmaking/internal/server"
	"matchmaking/pkg/grpc"
	"matchmaking/pkg/logger"
	//_ "net/http/pprof"
	"os"
	"os/signal"
)

// main is the entry point of service
func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	config, err := app.NewConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	// logger and metrics
	logger := logger.NewLogger(config.LogLevel)
	prometheusRegister := prometheusclient.DefaultRegisterer
	metrics.RegisterOn(prometheusRegister)
	defer metrics.UnRegisterFrom(prometheusRegister)

	// matchmaking service
	storage := matchmaking.NewStorage()
	service := matchmaking.NewService(logger, config.MatchmakingConfig, storage)
	matchOutput := service.Run(ctx)

	// grpc server
	matchmakingServer := server.NewMatchmakingServer(logger, service)
	grpcServer := grpc.NewGRPC(logger, config.PublicGrpcConfig).
		AddGrpcHealthCheck().
		AddServerImplementation(matchmakingServer.Register())

	// private API metrics
	privateApiBuilder := api.NewPrivateApi(logger, config.PrivateApiConfig)
	privateApiBuilder.RegisterPrivateRoutes()
	privateApi := privateApiBuilder.Build()

	// run servers
	group, errCtx := errgroup.WithContext(ctx)
	group.Go(func() error { return matchmakingServer.RunStatusUpdater(errCtx, matchOutput) })
	group.Go(func() error {
		return privateApi.ListenAndServe()
	})
	group.Go(func() error {
		return grpcServer.ListenAndServe()
	})

	// graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-errCtx.Done():
		logger.ErrorContext(errCtx, "failed to start private API", slog.String("error", errCtx.Err().Error()))
	case <-interrupt:
		logger.InfoContext(ctx, "shutting down")
	}
}
