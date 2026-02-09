module github.com/klemanjar0/payment-system/services/transaction

go 1.25.3

require (
    github.com/klemanjar0/payment-system/pkg v0.0.0
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.32.0
    github.com/google/uuid v1.5.0
    github.com/jackc/pgx/v5 v5.5.0
)

replace github.com/klemanjar0/payment-system/pkg => ../../pkg