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

	pb "github.com/klemanjar0/payment-system/generated/proto/transaction"
	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/config"
	"github.com/klemanjar0/payment-system/pkg/fiberutil"
	fiberMiddleware "github.com/klemanjar0/payment-system/pkg/fiberutil/middleware"
	"github.com/klemanjar0/payment-system/pkg/grpcutil"
	"github.com/klemanjar0/payment-system/pkg/httputil"
	pkgkafka "github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	pkgpostgres "github.com/klemanjar0/payment-system/pkg/postgres"
	"github.com/klemanjar0/payment-system/services/transaction/internal/client"
	grpcdelivery "github.com/klemanjar0/payment-system/services/transaction/internal/delivery/grpc"
	httpdelivery "github.com/klemanjar0/payment-system/services/transaction/internal/delivery/http"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
	txkafka "github.com/klemanjar0/payment-system/services/transaction/internal/kafka"
	pgrepository "github.com/klemanjar0/payment-system/services/transaction/internal/repository/postgres"
	usecase "github.com/klemanjar0/payment-system/services/transaction/internal/use_case"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// noopCommandPublisher is used when Kafka is not configured.
type noopCommandPublisher struct{}

func (n *noopCommandPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }

// noopBlacklist never blacklists tokens (transaction service has no Redis).
type noopBlacklist struct{}

func (b *noopBlacklist) Blacklist(_ context.Context, _ string, _ time.Duration) error { return nil }
func (b *noopBlacklist) IsBlacklisted(_ context.Context, _ string) (bool, error)      { return false, nil }

func main() {
	ctx := context.Background()

	isDev := config.GetEnvBool("DEV_MODE", true)
	logger.Init("transaction", isDev)

	// --- PostgreSQL ---
	pgPool, err := pkgpostgres.NewPool(ctx, pkgpostgres.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5435),
		User:     config.GetEnv("DB_USER", "transaction_service"),
		Password: config.GetEnv("DB_PASSWORD", "transaction_service_pass"),
		DBName:   config.GetEnv("DB_NAME", "transaction_db"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		MaxConns: int32(config.GetEnvInt("DB_MAX_CONNS", 10)),
	})
	if err != nil {
		logger.Fatal("failed to connect to postgres", "err", err)
	}
	defer pgPool.Close()

	// --- gRPC client: account service ---
	accountAddr := config.GetEnv("ACCOUNT_SERVICE_ADDR", "localhost:50052")
	accountConn, err := grpc.NewClient(accountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect to account service", "err", err, "addr", accountAddr)
	}
	defer accountConn.Close()
	accountClient := client.NewAccountClient(accountConn)

	// --- gRPC client: user service ---
	userAddr := config.GetEnv("USER_SERVICE_ADDR", "localhost:50051")
	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect to user service", "err", err, "addr", userAddr)
	}
	defer userConn.Close()
	userClient := client.NewUserClient(userConn)

	// --- Kafka command publisher ---
	var cmdPub usecase.CommandPublisher
	kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "")
	var kafkaCmdProducer *pkgkafka.Producer
	if kafkaBrokers != "" {
		kafkaCmdProducer = pkgkafka.NewProducer(pkgkafka.ProducerConfig{
			Brokers: strings.Split(kafkaBrokers, ","),
			Topic:   config.GetEnv("KAFKA_TRANSACTION_COMMANDS_TOPIC", "transaction-commands"),
		})
		defer kafkaCmdProducer.Close()
		cmdPub = txkafka.NewPublisher(kafkaCmdProducer)
	} else {
		logger.Info("KAFKA_BROKERS not set, using noop command publisher")
		cmdPub = &noopCommandPublisher{}
	}

	// --- Repository ---
	txRepo := pgrepository.NewTransactionRepository(pgPool)

	// --- Use cases ---
	createTransferUC := usecase.NewCreateTransferUseCase(txRepo, accountClient, userClient, cmdPub)
	getTransactionUC := usecase.NewGetTransactionUseCase(txRepo)
	getTransactionsByAccountUC := usecase.NewGetTransactionsByAccountUseCase(txRepo)

	// --- JWT token validator (verify-only, uses public key) ---
	publicKeyPath := config.GetEnv("PUBLIC_KEY_PATH", "keys/public-key.pem")
	publicKey, err := auth.LoadPublicKeyFromFile(publicKeyPath)
	if err != nil {
		logger.Fatal("failed to load public key", "err", err, "path", publicKeyPath)
	}
	tokenValidator, err := auth.NewTokenValidator(auth.Config{PublicKey: publicKey})
	if err != nil {
		logger.Fatal("failed to create token validator", "err", err)
	}

	// --- HTTP error mappings ---
	httputil.RegisterErrorMapping(domain.ErrTransactionNotFound, http.StatusNotFound)
	httputil.RegisterErrorMapping(domain.ErrTransactionExists, http.StatusConflict)
	httputil.RegisterErrorMapping(domain.ErrSameAccount, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrInvalidAmount, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrInvalidCurrency, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrAccountNotFound, http.StatusNotFound)
	httputil.RegisterErrorMapping(domain.ErrAccountNotActive, http.StatusUnprocessableEntity)
	httputil.RegisterErrorMapping(domain.ErrCurrencyMismatch, http.StatusUnprocessableEntity)
	httputil.RegisterErrorMapping(domain.ErrUserNotActive, http.StatusForbidden)

	// --- HTTP server (Fiber) ---
	app := fiberutil.NewApp()
	app.Use(fiberMiddleware.Recovery())
	app.Use(fiberMiddleware.Logging())

	httpHandler := httpdelivery.NewTransactionHTTPHandler(createTransferUC, getTransactionUC, getTransactionsByAccountUC)
	httpdelivery.RegisterRoutes(app, httpHandler, fiberMiddleware.Auth(tokenValidator, &noopBlacklist{}))

	httpPort := config.GetEnv("HTTP_PORT", "8082")

	// --- gRPC server ---
	grpcPort := config.GetEnv("GRPC_PORT", "50053")
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

	grpcHandler := grpcdelivery.NewServer(createTransferUC, getTransactionUC, getTransactionsByAccountUC)
	pb.RegisterTransactionServiceServer(srv, grpcHandler)

	// --- Kafka Saga orchestrator ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if kafkaBrokers != "" {
		accountEventsConsumer := pkgkafka.NewConsumer(pkgkafka.ConsumerConfig{
			Brokers: strings.Split(kafkaBrokers, ","),
			Topic:   config.GetEnv("KAFKA_ACCOUNT_EVENTS_TOPIC", "account-events"),
			GroupID: config.GetEnv("KAFKA_GROUP_ID", "transaction-service"),
		})
		defer accountEventsConsumer.Close()

		orchestrator := txkafka.NewSagaOrchestrator(
			accountEventsConsumer,
			txRepo,
			txkafka.NewPublisher(kafkaCmdProducer),
		)

		go func() {
			orchestrator.Run(ctx)
		}()
	} else {
		logger.Info("KAFKA_BROKERS not set, saga orchestrator disabled")
	}

	go func() {
		logger.Info("transaction grpc service started", "port", grpcPort)
		if err := srv.Serve(lis); err != nil {
			logger.Fatal("grpc server failed", "err", err)
		}
	}()

	go func() {
		logger.Info("transaction http service started", "port", httpPort)
		if err := app.Listen(":" + httpPort); err != nil {
			logger.Fatal("http server failed", "err", err)
		}
	}()

	<-quit
	logger.Info("shutting down transaction service")
	srv.GracefulStop()
	_ = app.Shutdown()
}
