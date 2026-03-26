FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod/go.sum for all workspace modules first (for layer caching of go mod download)
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY generated/proto/go.mod generated/proto/go.sum ./generated/proto/
COPY services/user/go.mod services/user/go.sum ./services/user/
COPY services/account/go.mod services/account/go.sum ./services/account/
COPY services/transaction/go.mod services/transaction/go.sum ./services/transaction/

RUN cd services/user && go mod download

# Copy full source
COPY pkg/ ./pkg/
COPY generated/ ./generated/
COPY services/user/ ./services/user/
COPY keys/ ./keys/

RUN cd services/user && CGO_ENABLED=0 go build -o /bin/user ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/user /bin/user
COPY services/user/migrations /migrations

EXPOSE 50051

ENTRYPOINT ["/bin/user"]
