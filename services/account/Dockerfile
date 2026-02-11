FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY pkg/ ./pkg/
COPY services/account/ ./services/account/
COPY generated/ ./generated/
COPY go.work go.work.sum ./

RUN cd services/account && go mod download
RUN cd services/account && CGO_ENABLED=0 go build -o /bin/account ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/account /bin/account
COPY services/account/migrations /migrations

EXPOSE 50052

ENTRYPOINT ["/bin/account"]
