package grpcutil

import (
	"context"
	"time"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := codes.OK
		if err != nil {
			code = status.Code(err)
		}

		logger.Info("grpc request",
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"code", code.String(),
		)

		return resp, err
	}
}

func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					"method", info.FullMethod,
					"panic", r,
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
