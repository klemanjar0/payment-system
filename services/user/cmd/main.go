package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "github.com/klemanjar0/payment-system/generated/proto/user"
	pkgauditlog "github.com/klemanjar0/payment-system/pkg/auditlog"
	pgauditlog "github.com/klemanjar0/payment-system/pkg/auditlog/postgres"
	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/config"
	"github.com/klemanjar0/payment-system/pkg/fiberutil"
	fiberMiddleware "github.com/klemanjar0/payment-system/pkg/fiberutil/middleware"
	"github.com/klemanjar0/payment-system/pkg/grpcutil"
	"github.com/klemanjar0/payment-system/pkg/httputil"
	"github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	pkgpostgres "github.com/klemanjar0/payment-system/pkg/postgres"
	pkgredis "github.com/klemanjar0/payment-system/pkg/redis"
	userauditlog "github.com/klemanjar0/payment-system/services/user/internal/auditlog"
	grpcdelivery "github.com/klemanjar0/payment-system/services/user/internal/delivery/grpc"
	httpdelivery "github.com/klemanjar0/payment-system/services/user/internal/delivery/http"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
	"github.com/klemanjar0/payment-system/services/user/internal/repository/cached"
	pgrepository "github.com/klemanjar0/payment-system/services/user/internal/repository/postgres"
	"github.com/klemanjar0/payment-system/services/user/internal/usecase"
	"google.golang.org/grpc"
)

// kafkaEventPublisher adapts kafka.Producer to the usecase.EventPublisher interface.
type kafkaEventPublisher struct {
	producer *kafka.Producer
}

func (k *kafkaEventPublisher) Publish(ctx context.Context, eventType string, payload any) error {
	return k.producer.Publish(ctx, eventType, kafka.Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	})
}

// noopEventPublisher is used when Kafka is not configured.
type noopEventPublisher struct{}

func (n *noopEventPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }

func main() {
	ctx := context.Background()

	isDev := config.GetEnvBool("DEV_MODE", true)
	logger.Init("user", isDev)

	// --- PostgreSQL ---
	pgPool, err := pkgpostgres.NewPool(ctx, pkgpostgres.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5433),
		User:     config.GetEnv("DB_USER", "user_service"),
		Password: config.GetEnv("DB_PASSWORD", "user_service_pass"),
		DBName:   config.GetEnv("DB_NAME", "user_db"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		MaxConns: int32(config.GetEnvInt("DB_MAX_CONNS", 10)),
	})
	if err != nil {
		logger.Fatal("failed to connect to postgres", "err", err)
	}

	// --- Redis ---
	redisClient, err := pkgredis.NewClient(ctx, pkgredis.Config{
		Host:     config.GetEnv("REDIS_HOST", "localhost"),
		Port:     config.GetEnvInt("REDIS_PORT", 6379),
		Password: config.GetEnv("REDIS_PASSWORD", ""),
		DB:       config.GetEnvInt("REDIS_DB", 0),
	})
	if err != nil {
		logger.Fatal("failed to connect to redis", "err", err)
	}

	// --- Token service ---
	privateKeyPath := config.GetEnv("PRIVATE_KEY_PATH", "keys/private.pem")
	privateKey, err := auth.LoadPrivateKeyFromFile(privateKeyPath)
	if err != nil {
		logger.Fatal("failed to load private key", "err", err, "path", privateKeyPath)
	}

	tokenSvc, err := auth.NewTokenService(auth.Config{
		PrivateKey: privateKey,
	})
	if err != nil {
		logger.Fatal("failed to create token service", "err", err)
	}

	// --- Kafka event publisher ---
	var eventPub usecase.EventPublisher
	if kafkaBrokers := config.GetEnv("KAFKA_BROKERS", ""); kafkaBrokers != "" {
		eventPub = &kafkaEventPublisher{
			producer: kafka.NewProducer(kafka.ProducerConfig{
				Brokers: strings.Split(kafkaBrokers, ","),
				Topic:   "user-events",
			}),
		}
	} else {
		eventPub = &noopEventPublisher{}
	}

	// --- Repositories ---
	userRepo := pgrepository.NewUserRepository(pgPool)
	cachedUserRepository := cached.NewCachedUserRepository(userRepo, redisClient)

	// --- Audit log stack:
	//   pgauditlog.Repository → pkgauditlog.Logger → userauditlog.UserAuditLogger
	auditRepo := pgauditlog.New(pgPool)
	auditLogger := pkgauditlog.New(auditRepo, "user")
	userAudit := userauditlog.New(auditLogger)

	// --- Use cases ---
	createUserUC := usecase.NewCreateUserUseCase(userRepo, tokenSvc, eventPub, userAudit)
	authenticateUC := usecase.NewAuthenticateUseCase(userRepo, tokenSvc, userAudit)
	getUserUC := usecase.NewGetUserUseCase(cachedUserRepository)
	changePasswordUC := usecase.NewChangePasswordUseCase(userRepo, userAudit)

	// --- HTTP error mappings ---
	httputil.RegisterErrorMapping(domain.ErrUserNotFound, http.StatusNotFound)
	httputil.RegisterErrorMapping(domain.ErrUserAlreadyExists, http.StatusConflict)
	httputil.RegisterErrorMapping(domain.ErrInvalidCredentials, http.StatusUnauthorized)
	httputil.RegisterErrorMapping(domain.ErrUserBlocked, http.StatusForbidden)
	httputil.RegisterErrorMapping(domain.ErrUserNotActive, http.StatusForbidden)
	httputil.RegisterErrorMapping(domain.ErrInvalidEmail, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrInvalidPhone, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrPasswordTooShort, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrPasswordTooLong, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrPasswordTooWeak, http.StatusBadRequest)

	// --- HTTP server (Fiber) ---
	app := fiberutil.NewApp()
	app.Use(fiberMiddleware.Recovery())
	app.Use(fiberMiddleware.Logging())

	httpHandler := httpdelivery.NewUserHTTPHandler(createUserUC, authenticateUC, getUserUC, changePasswordUC)
	httpdelivery.RegisterRoutes(app, httpHandler, fiberMiddleware.Auth(tokenSvc))

	httpPort := config.GetEnv("HTTP_PORT", "8080")

	// --- gRPC server ---
	grpcPort := config.GetEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Fatal("failed to listen", "err", err, "port", grpcPort)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcutil.RecoveryInterceptor(),
			grpcutil.LoggingInterceptor(),
		),
	)

	handler := grpcdelivery.NewUserHandler(createUserUC, authenticateUC, getUserUC, changePasswordUC)
	pb.RegisterUserServiceServer(srv, handler)

	// --- Graceful shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("user grpc service started", "port", grpcPort)
		if err := srv.Serve(lis); err != nil {
			logger.Fatal("grpc server failed", "err", err)
		}
	}()

	go func() {
		logger.Info("user http service started", "port", httpPort)
		if err := app.Listen(":" + httpPort); err != nil {
			logger.Fatal("http server failed", "err", err)
		}
	}()

	<-quit
	logger.Info("shutting down user service")
	srv.GracefulStop()
	_ = app.Shutdown()
}
