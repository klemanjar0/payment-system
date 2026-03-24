package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "github.com/klemanjar0/payment-system/generated/proto/user"
	pkgauditlog "github.com/klemanjar0/payment-system/pkg/auditlog"
	mongoauditlog "github.com/klemanjar0/payment-system/pkg/auditlog/mongodb"
	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/config"
	"github.com/klemanjar0/payment-system/pkg/grpcutil"
	"github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/pkg/mongodb"
	pkgpostgres "github.com/klemanjar0/payment-system/pkg/postgres"
	pkgredis "github.com/klemanjar0/payment-system/pkg/redis"
	userauditlog "github.com/klemanjar0/payment-system/services/user/internal/auditlog"
	grpcdelivery "github.com/klemanjar0/payment-system/services/user/internal/delivery/grpc"
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

	// --- MongoDB (audit logs) ---
	auditDB := config.GetEnv("MONGO_AUDIT_DB", "audit")
	mongoClient, err := mongodb.NewClient(ctx, mongodb.Config{
		URI:      config.GetEnv("MONGO_URI", "mongodb://localhost:27017"),
		Database: auditDB,
	})
	if err != nil {
		logger.Fatal("failed to connect to mongodb", "err", err)
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

	// --- Audit log stack (3 layers):
	//   mongoauditlog.Repository → pkgauditlog.Logger → userauditlog.UserAuditLogger
	auditRepo := mongoauditlog.New(mongoClient, auditDB, "audit_logs")

	if err := auditRepo.EnsureIndexes(ctx); err != nil {
		logger.Warn("failed to ensure audit log indexes", "err", err)
	}

	auditLogger := pkgauditlog.New(auditRepo, "user")
	userAudit := userauditlog.New(auditLogger)

	// --- Use cases ---
	createUserUC := usecase.NewCreateUserUseCase(userRepo, tokenSvc, eventPub, userAudit)
	authenticateUC := usecase.NewAuthenticateUseCase(userRepo, tokenSvc, userAudit)
	getUserUC := usecase.NewGetUserUseCase(cachedUserRepository)
	changePasswordUC := usecase.NewChangePasswordUseCase(userRepo, userAudit)

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
		logger.Info("user service started", "port", grpcPort)
		if err := srv.Serve(lis); err != nil {
			logger.Fatal("grpc server failed", "err", err)
		}
	}()

	<-quit
	logger.Info("shutting down user service")
	srv.GracefulStop()
}
