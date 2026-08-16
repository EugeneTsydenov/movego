FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
COPY gen/go.mod gen/go.sum ./gen/

RUN go mod download

COPY . .

WORKDIR /app/services/user-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server ./cmd/app

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/services/user-service/configs /app/configs

EXPOSE 8080
EXPOSE 8081
EXPOSE 2112

CMD ["/app/server", "--config=configs"]
