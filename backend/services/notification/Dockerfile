FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY services/notification/go.mod services/notification/go.sum ./services/notification/
COPY services/user/go.mod services/user/go.sum ./services/user/
COPY services/game/go.mod services/game/go.sum ./services/game/
COPY services/matchmaking/go.mod services/matchmaking/go.sum ./services/matchmaking/
COPY gen/go.mod gen/go.sum ./gen/
COPY shared/go.mod ./shared/

RUN go mod download

COPY . .

WORKDIR /app/services/notification
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server ./cmd/app

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/services/notification/configs /app/configs

EXPOSE 1000
EXPOSE 2132

CMD ["/app/server", "--config=configs"]
