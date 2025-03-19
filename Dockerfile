FROM docker.io/golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/prometheus-proxy .
#-----

FROM docker.io/alpine:3.18

WORKDIR /app

COPY --from=builder /app/bin/prometheus-proxy /app/prometheus-proxy

COPY config.yaml /app/config.yaml

CMD ["/app/prometheus-proxy"]
