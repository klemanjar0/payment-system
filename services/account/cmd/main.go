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

	pb "github.com/klemanjar0/payment-system/generated/proto/account"
	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/config"
	"github.com/klemanjar0/payment-system/pkg/fiberutil"
	fiberMiddleware "github.com/klemanjar0/payment-system/pkg/fiberutil/middleware"
	"github.com/klemanjar0/payment-system/pkg/grpcutil"
	"github.com/klemanjar0/payment-system/pkg/httputil"
	pkgkafka "github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	pkgpostgres "github.com/klemanjar0/payment-system/pkg/postgres"
	grpcdelivery "github.com/klemanjar0/payment-system/services/account/internal/delivery/grpc"
	httpdelivery "github.com/klemanjar0/payment-system/services/account/internal/delivery/http"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	acckafka "github.com/klemanjar0/payment-system/services/account/internal/kafka"
	pgrepository "github.com/klemanjar0/payment-system/services/account/internal/repository/postgres"
	usecase "github.com/klemanjar0/payment-system/services/account/internal/use_case"
	"google.golang.org/grpc"
)

// noopEventPublisher is used when Kafka is not configured.
type noopEventPublisher struct{}

func (n *noopEventPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }

// noopBlacklist never blacklists tokens (account service has no Redis).
type noopBlacklist struct{}

func (b *noopBlacklist) Blacklist(_ context.Context, _ string, _ time.Duration) error { return nil }
func (b *noopBlacklist) IsBlacklisted(_ context.Context, _ string) (bool, error)      { return false, nil }

func main() {
	ctx := context.Background()

	isDev := config.GetEnvBool("DEV_MODE", true)
	logger.Init("account", isDev)

	// --- PostgreSQL ---
	pgPool, err := pkgpostgres.NewPool(ctx, pkgpostgres.Config{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnvInt("DB_PORT", 5434),
		User:     config.GetEnv("DB_USER", "account_service"),
		Password: config.GetEnv("DB_PASSWORD", "account_service_pass"),
		DBName:   config.GetEnv("DB_NAME", "account_db"),
		SSLMode:  config.GetEnv("DB_SSL_MODE", "disable"),
		MaxConns: int32(config.GetEnvInt("DB_MAX_CONNS", 10)),
	})
	if err != nil {
		logger.Fatal("failed to connect to postgres", "err", err)
	}
	defer pgPool.Close()

	// --- Kafka event publisher ---
	var eventPub usecase.EventPublisher
	kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "")
	if kafkaBrokers != "" {
		producer := pkgkafka.NewProducer(pkgkafka.ProducerConfig{
			Brokers: strings.Split(kafkaBrokers, ","),
			Topic:   config.GetEnv("KAFKA_ACCOUNT_EVENTS_TOPIC", "account-events"),
		})
		defer producer.Close()
		eventPub = acckafka.NewPublisher(producer)
	} else {
		logger.Info("KAFKA_BROKERS not set, using noop event publisher")
		eventPub = &noopEventPublisher{}
	}

	// --- Repositories ---
	accountRepo := pgrepository.NewAccountRepository(pgPool)
	holdRepo := pgrepository.NewHoldRepository(pgPool)
	operationRepo := pgrepository.NewOperationRepository(pgPool)
	txManager := pgrepository.NewTxManager(pgPool)

	_ = holdRepo      // used via TxManager inside transactions
	_ = operationRepo // used via TxManager inside transactions

	// --- Use cases ---
	createAccountUC := usecase.NewCreateAccountUseCase(accountRepo)
	getAccountUC := usecase.NewGetAccountUseCase(accountRepo)
	getAccountsByUserUC := usecase.NewGetAccountsByUserUseCase(accountRepo)

	placeHoldUC := usecase.NewPlaceHoldUseCase(txManager, eventPub)
	executeHoldUC := usecase.NewExecuteHoldUseCase(txManager, eventPub)
	releaseHoldUC := usecase.NewReleaseHoldUseCase(txManager, eventPub)
	creditUC := usecase.NewCreditUseCase(txManager, eventPub)

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
	httputil.RegisterErrorMapping(domain.ErrAccountNotFound, http.StatusNotFound)
	httputil.RegisterErrorMapping(domain.ErrAccountExists, http.StatusConflict)
	httputil.RegisterErrorMapping(domain.ErrInvalidCurrency, http.StatusBadRequest)
	httputil.RegisterErrorMapping(domain.ErrAccountNotActive, http.StatusForbidden)
	httputil.RegisterErrorMapping(domain.ErrInsufficientFunds, http.StatusUnprocessableEntity)

	// --- HTTP server (Fiber) ---
	app := fiberutil.NewApp()
	app.Use(fiberMiddleware.Recovery())
	app.Use(fiberMiddleware.Logging())

	httpHandler := httpdelivery.NewAccountHTTPHandler(createAccountUC, getAccountUC, getAccountsByUserUC)
	httpdelivery.RegisterRoutes(app, httpHandler, fiberMiddleware.Auth(tokenValidator, &noopBlacklist{}))

	httpPort := config.GetEnv("HTTP_PORT", "8081")

	// --- gRPC server ---
	grpcPort := config.GetEnv("GRPC_PORT", "50052")
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

	grpcHandler := grpcdelivery.NewServer(createAccountUC, getAccountUC, getAccountsByUserUC)
	pb.RegisterAccountServiceServer(srv, grpcHandler)

	// --- Kafka Saga consumer ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	if kafkaBrokers != "" {
		sagaConsumer := pkgkafka.NewConsumer(pkgkafka.ConsumerConfig{
			Brokers: strings.Split(kafkaBrokers, ","),
			Topic:   config.GetEnv("KAFKA_TRANSACTION_COMMANDS_TOPIC", "transaction-commands"),
			GroupID: config.GetEnv("KAFKA_GROUP_ID", "account-service"),
		})
		defer sagaConsumer.Close()

		sagaHandler := acckafka.NewSagaConsumer(sagaConsumer, placeHoldUC, executeHoldUC, releaseHoldUC, creditUC)

		go func() {
			sagaHandler.Run(ctx)
		}()
	} else {
		logger.Info("KAFKA_BROKERS not set, saga consumer disabled")
	}

	go func() {
		logger.Info("account grpc service started", "port", grpcPort)
		if err := srv.Serve(lis); err != nil {
			logger.Fatal("grpc server failed", "err", err)
		}
	}()

	go func() {
		logger.Info("account http service started", "port", httpPort)
		if err := app.Listen(":" + httpPort); err != nil {
			logger.Fatal("http server failed", "err", err)
		}
	}()

	<-quit
	logger.Info("shutting down account service")
	srv.GracefulStop()
	_ = app.Shutdown()

}
