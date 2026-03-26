FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod/go.sum for all workspace modules first (for layer caching of go mod download)
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY generated/proto/go.mod generated/proto/go.sum ./generated/proto/
COPY services/transaction/go.mod services/transaction/go.sum ./services/transaction/
COPY services/account/go.mod services/account/go.sum ./services/account/
COPY services/user/go.mod services/user/go.sum ./services/user/

RUN cd services/transaction && go mod download

# Copy full source (transaction imports services/account/events)
COPY pkg/ ./pkg/
COPY generated/ ./generated/
COPY services/transaction/ ./services/transaction/
COPY services/account/ ./services/account/

RUN cd services/transaction && CGO_ENABLED=0 go build -o /bin/transaction ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/transaction /bin/transaction
COPY services/transaction/migrations /migrations

EXPOSE 50053

ENTRYPOINT ["/bin/transaction"]
