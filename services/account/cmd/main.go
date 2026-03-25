package cmd

import (
	"context"

	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/config"
	"github.com/klemanjar0/payment-system/pkg/logger"
	pkgpostgres "github.com/klemanjar0/payment-system/pkg/postgres"
)

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

	// --- Token service ---
	publicKeyPath := config.GetEnv("PUBLIC_KEY_PATH", "keys/public.pem")
	publicKey, err := auth.LoadPublicKeyFromFile(publicKeyPath)
	if err != nil {
		logger.Fatal("failed to load publicKey key", "err", err, "path", publicKeyPath)
	}

	tokenSvc, err := auth.NewTokenValidator(auth.Config{
		PublicKey: publicKey,
	})
	if err != nil {
		logger.Fatal("failed to create token validator", "err", err)
	}
}
