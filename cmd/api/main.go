package main

import (
	"flag"
	_ "github.com/lib/pq"
	"os"
	"os/signal"
	"strconv"
	"subscriptionMService/internal/app/grpcapp"
	"subscriptionMService/internal/jsonlog"
	"subscriptionMService/internal/planCache"
	"subscriptionMService/internal/services/subscription"
	"subscriptionMService/storage/postgres"
	"syscall"
	"time"
)

const version = "1.0.0"

type StorageDetails struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type Config struct {
	env      string
	DB       StorageDetails
	GRPC     GRPCConfig
	TokenTTL time.Duration
}

type GRPCConfig struct {
	Port    int
	Timeout time.Duration
}

type Application struct {
	GRPCSrv *grpcapp.App
}

func main() {
	var cfg Config

	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.DB.DSN, "db-dsn", "postgres://sub:pass@localhost:5432/subscriptions?sslmode=disable&client_encoding=UTF8", "PostgresSQL DSN")
	flag.IntVar(&cfg.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgresSQL max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "db-max-Idle-conns", 25, "PostgresSQL max Idle connections")
	flag.StringVar(&cfg.DB.MaxIdleTime, "db-max-Idle-time", "15m", "PostgresSQl max Idle time")

	flag.IntVar(&cfg.GRPC.Port, "grpc-port", 3000, "grpc-port")
	flag.DurationVar(&cfg.TokenTTL, "token-ttl", time.Hour, "GRPC's work duration")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	app := New(logger, cfg.GRPC.Port, cfg, cfg.TokenTTL)

	logger.PrintInfo("connection pool established", map[string]string{
		"port": strconv.Itoa(cfg.GRPC.Port),
	})
	go app.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	logger.PrintInfo("stopping application", map[string]string{
		"signal": sign.String(),
	})

	app.GRPCSrv.Stop()
}

func New(log *jsonlog.Logger, grpcPort int, cfg Config, tokenTTL time.Duration) *Application {
	dbcfg := postgres.StorageDetails(cfg.DB)
	db, err := postgres.OpenDB(dbcfg)
	if err != nil {
		log.PrintFatal(err, nil)
	}

	defer db.Close()

	planCacheProvider := planCache.NewCachedPlanProvider(db, tokenTTL)

	subscriptionService := subscription.New(log, db, planCacheProvider, tokenTTL)
	grpcApp := grpcapp.New(log, grpcPort, subscriptionService) // добавить сервис

	return &Application{GRPCSrv: grpcApp}
}
