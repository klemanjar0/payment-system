package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	pb "github.com/klemanjar0/payment-system/generated/proto/account"
	"github.com/klemanjar0/payment-system/pkg/config"
	"github.com/klemanjar0/payment-system/pkg/grpcutil"
	pkgkafka "github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	pkgpostgres "github.com/klemanjar0/payment-system/pkg/postgres"
	grpcdelivery "github.com/klemanjar0/payment-system/services/account/internal/delivery/grpc"
	kafkadeliv "github.com/klemanjar0/payment-system/services/account/internal/kafka"
	kafkapub "github.com/klemanjar0/payment-system/services/account/internal/kafka"
	pgrepository "github.com/klemanjar0/payment-system/services/account/internal/repository/postgres"
	usecase "github.com/klemanjar0/payment-system/services/account/internal/use_case"
	"google.golang.org/grpc"
)

// noopEventPublisher is used when Kafka is not configured.
type noopEventPublisher struct{}

func (n *noopEventPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }

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
		eventPub = kafkapub.NewPublisher(producer)
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

		sagaHandler := kafkadeliv.NewSagaConsumer(sagaConsumer, placeHoldUC, executeHoldUC, releaseHoldUC, creditUC)

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

	<-quit
	logger.Info("shutting down account service")
	srv.GracefulStop()
}
