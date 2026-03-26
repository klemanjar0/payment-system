module github.com/klemanjar0/payment-system/services/account

go 1.25.3

require (
	github.com/jackc/pgx/v5 v5.9.1
	github.com/klemanjar0/payment-system/generated/proto v0.0.0
	github.com/klemanjar0/payment-system/pkg v0.0.0
	github.com/klemanjar0/payment-system/services/transaction v0.0.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.50 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
)

replace github.com/klemanjar0/payment-system/pkg => ../../pkg

replace github.com/klemanjar0/payment-system/generated/proto => ../../generated/proto

replace github.com/klemanjar0/payment-system/services/transaction => ../transaction
