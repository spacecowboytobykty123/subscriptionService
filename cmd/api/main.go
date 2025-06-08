package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	subs "github.com/spacecowboytobykty123/subsProto/gen/go/subscription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"subscriptionMService/internal/app/grpcapp"
	bcktgrpc "subscriptionMService/internal/clients/bucket/grpc"
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
	Clients  ClientsConfig
}

type Client struct {
	Address      int           `yaml:"address"`
	Timeout      time.Duration `yaml:"timeout"`
	RetriesCount int           `yaml:"retries_count"`
	insecure     bool          `yaml:"insecure"`
}

type ClientsConfig struct {
	Bucket Client `yaml:"bucket"`
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
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&client_encoding=UTF8", user, pass, host, port, name)

	flag.StringVar(&cfg.DB.DSN, "db-dsn", dsn, "PostgresSQL DSN")
	flag.IntVar(&cfg.DB.MaxOpenConns, "db-max-open-conns", 25, "PostgresSQL max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "db-max-Idle-conns", 25, "PostgresSQL max Idle connections")
	flag.StringVar(&cfg.DB.MaxIdleTime, "db-max-Idle-time", "15m", "PostgresSQl max Idle time")

	flag.IntVar(&cfg.Clients.Bucket.Address, "bucket-client-addr", 2000, "bucket-port")
	flag.IntVar(&cfg.GRPC.Port, "grpc-port", 3000, "grpc-port")
	flag.DurationVar(&cfg.TokenTTL, "token-ttl", time.Hour, "GRPC's work duration")

	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	bucketClient, err := bcktgrpc.New(context.Background(), logger, cfg.Clients.Bucket.Timeout, cfg.Clients.Bucket.Address)
	if err != nil {
		logger.PrintError(err, map[string]string{
			"message": "failed to init bucket client",
		})
	}
	app := New(logger, cfg.GRPC.Port, cfg, cfg.TokenTTL, bucketClient)

	logger.PrintInfo("connection pool established", map[string]string{
		"port": strconv.Itoa(cfg.GRPC.Port),
	})
	go app.GRPCSrv.MustRun()
	go runHTTP(cfg.GRPC.Port, logger)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	logger.PrintInfo("stopping application", map[string]string{
		"signal": sign.String(),
	})

	app.GRPCSrv.Stop()
}

func New(log *jsonlog.Logger, grpcPort int, cfg Config, tokenTTL time.Duration, bucketClient *bcktgrpc.BucketClient) *Application {
	dbcfg := postgres.StorageDetails(cfg.DB)
	db, err := postgres.OpenDB(dbcfg)
	if err != nil {
		log.PrintFatal(err, nil)
	}

	//defer db.Close()

	planCacheProvider := planCache.NewCachedPlanProvider(db, tokenTTL)

	subscriptionService := subscription.New(log, db, planCacheProvider, bucketClient, tokenTTL)
	grpcApp := grpcapp.New(log, grpcPort, subscriptionService) // добавить сервис

	return &Application{GRPCSrv: grpcApp}
}

func runHTTP(grpcPort int, logger *jsonlog.Logger) {
	ctx := context.Background()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	endpoint := "localhost:" + strconv.Itoa(grpcPort)
	if err := subs.RegisterSubscriptionHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.PrintFatal(err, map[string]string{
			"method":  "main.runHTTP",
			"message": "failed to start HTTP gateway",
		})
	}
	fs := http.FileServer(http.Dir("C:\\Users\\Еркебулан\\GolandProjects\\subsProto\\gen\\swagger")) // path where swagger.json is output
	http.Handle("/swagger/", http.StripPrefix("/swagger/", fs))
	http.Handle("/", mux)

	logger.PrintInfo("HTTP REST gateway and Swagger docs started", map[string]string{
		"port": "9090",
	})

	if err := http.ListenAndServe(":9090", mux); err != nil {
		logger.PrintFatal(err, map[string]string{
			"message": "could not start http server",
			"method":  "main.runHTTp",
		})
	}

}

//func runRest() {
//	ctx := context.Background()
//	ctx, cancel := context.WithCancel(ctx)
//	defer cancel()
//	mux := runtime.NewServeMux()
//	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
//	err :=
//	if err != nil {
//		panic(err)
//	}
//	log.Printf("server listening at 8081")
//	if err := http.ListenAndServe(":8081", mux); err != nil {
//		panic(err)
//	}
//}
