FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY pkg/ ./pkg/
COPY services/transaction/ ./services/transaction/
COPY generated/ ./generated/
COPY go.work go.work.sum ./

RUN cd services/transaction && go mod download
RUN cd services/transaction && CGO_ENABLED=0 go build -o /bin/transaction ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/transaction /bin/transaction
COPY services/transaction/migrations /migrations

EXPOSE 50053

ENTRYPOINT ["/bin/transaction"]
