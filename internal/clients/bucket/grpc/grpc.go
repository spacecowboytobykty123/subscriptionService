package grpc

import (
	"context"
	"fmt"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	bckt "github.com/spacecowboytobykty123/bucketProto/gen/go/bucket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"subscriptionMService/internal/jsonlog"
	"time"
)

type BucketClient struct {
	bucketApi bckt.BucketClient
	log       *jsonlog.Logger
}

func New(ctx context.Context, log *jsonlog.Logger, timeout time.Duration, retriesCount int) (*BucketClient, error) {
	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadSent, grpclog.PayloadReceived),
	}

	cc, err := grpc.DialContext(ctx, "localhost:2000",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("%s:%w", "grpc.New", err)
	}
	return &BucketClient{
		bucketApi: bckt.NewBucketClient(cc),
		log:       log,
	}, nil
}

func InterceptorLogger(logger *jsonlog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		logger.PrintInfo(msg, map[string]string{
			"lvl": string(lvl),
		})
	},
	)
}

func (b *BucketClient) CreateBucket(ctx context.Context) *bckt.CreateBucketResponse {
	b.log.PrintInfo("creating toys in bucket service", map[string]string{
		"method":  "bucket.grpc.createBucket",
		"service": "bucket",
	})
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		b.log.PrintError(fmt.Errorf("missing metadata"), map[string]string{
			"method":  "bucket.grpc.getBucket",
			"service": "bucket",
		})
		return &bckt.CreateBucketResponse{
			Status: bckt.OperationStatus_STATUS_INTERNAL_ERROR,
			Msg:    "could not get create new metadata",
		}
	}
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		b.log.PrintError(fmt.Errorf("missing authorization token"), nil)
		return &bckt.CreateBucketResponse{
			Status: bckt.OperationStatus_STATUS_INTERNAL_ERROR,
			Msg:    "missing auth token",
		}
	}
	outctx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", authHeader[0]))
	b.log.PrintInfo("forwarding JWT token", map[string]string{
		"token": authHeader[0],
	})

	reps, err := b.bucketApi.CreateBucket(outctx, &bckt.CreateBucketRequest{})
	if err != nil {
		b.log.PrintError(fmt.Errorf("could not get response from bucket service"), map[string]string{
			"method":  "bucket.grpc.createBucket",
			"service": "bucket",
		})
	}
	return reps
}
