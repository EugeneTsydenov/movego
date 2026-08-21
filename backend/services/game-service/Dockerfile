FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY services/game-service/go.mod services/game-service/go.sum ./services/game-service/
COPY gen/go.mod gen/go.sum ./gen/
COPY shared/go.mod ./shared/

RUN go mod download

COPY . .

WORKDIR /app/services/game-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server ./cmd/app

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/services/game-service/configs /app/configs

EXPOSE 8082
EXPOSE 2113

CMD ["/app/server", "--config=configs"]
