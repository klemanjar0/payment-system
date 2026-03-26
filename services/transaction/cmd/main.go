package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	pb "github.com/klemanjar0/payment-system/generated/proto/transaction"
	"github.com/klemanjar0/payment-system/pkg/config"
	"github.com/klemanjar0/payment-system/pkg/grpcutil"
	pkgkafka "github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	pkgpostgres "github.com/klemanjar0/payment-system/pkg/postgres"
	"github.com/klemanjar0/payment-system/services/transaction/internal/client"
	grpcdelivery "github.com/klemanjar0/payment-system/services/transaction/internal/delivery/grpc"
	kafkalayer "github.com/klemanjar0/payment-system/services/transaction/internal/kafka"
	pgrepository "github.com/klemanjar0/payment-system/services/transaction/internal/repository/postgres"
	usecase "github.com/klemanjar0/payment-system/services/transaction/internal/use_case"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// noopCommandPublisher is used when Kafka is not configured.
type noopCommandPublisher struct{}

func (n *noopCommandPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }

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
		cmdPub = kafkalayer.NewPublisher(kafkaCmdProducer)
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

		orchestrator := kafkalayer.NewSagaOrchestrator(
			accountEventsConsumer,
			txRepo,
			kafkalayer.NewPublisher(kafkaCmdProducer),
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

	<-quit
	logger.Info("shutting down transaction service")
	srv.GracefulStop()
}
