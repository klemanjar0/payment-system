FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY pkg/ ./pkg/
COPY services/user/ ./services/user/
COPY generated/ ./generated/
COPY go.work go.work.sum ./

RUN cd services/user && go mod download
RUN cd services/user && CGO_ENABLED=0 go build -o /bin/user ./cmd

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/user /bin/user
COPY services/user/migrations /migrations

EXPOSE 50051

ENTRYPOINT ["/bin/user"]
