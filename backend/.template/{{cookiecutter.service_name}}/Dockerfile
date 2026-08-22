FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY services/{{cookiecutter.service_name}}/go.mod services/{{cookiecutter.service_name}}/go.sum ./services/{{cookiecutter.service_name}}/
COPY gen/go.mod gen/go.sum ./gen/
COPY shared/go.mod ./shared/

RUN go mod download

COPY . .

WORKDIR /app/services/{{cookiecutter.service_name}}
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server ./cmd/app

FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/services/{{cookiecutter.service_name}}/configs /app/configs

EXPOSE {{cookiecutter.port}}
EXPOSE {{cookiecutter.metrics_port}}

CMD ["/app/server", "--config=configs"]
