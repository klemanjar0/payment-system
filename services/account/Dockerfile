FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod/go.sum for all workspace modules first (for layer caching of go mod download)
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY generated/proto/go.mod generated/proto/go.sum ./generated/proto/
COPY services/account/go.mod services/account/go.sum ./services/account/
COPY services/transaction/go.mod services/transaction/go.sum ./services/transaction/
COPY services/user/go.mod services/user/go.sum ./services/user/

RUN cd services/account && go mod download

# Copy full source (account imports services/transaction/events and services/account/events)
COPY pkg/ ./pkg/
COPY generated/ ./generated/
COPY services/account/ ./services/account/
COPY services/transaction/ ./services/transaction/

RUN cd services/account && CGO_ENABLED=0 go build -o /bin/account ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/account /bin/account
COPY services/account/migrations /migrations

EXPOSE 50052

ENTRYPOINT ["/bin/account"]
