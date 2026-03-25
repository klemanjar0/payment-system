module github.com/klemanjar0/payment-system/services/account

go 1.25.3

require github.com/klemanjar0/payment-system/pkg v0.0.0

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/klemanjar0/payment-system/pkg => ../../pkg

replace github.com/klemanjar0/payment-system/generated/proto => ../../generated/proto
