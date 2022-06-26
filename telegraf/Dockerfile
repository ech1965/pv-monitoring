FROM golang:alpine AS builder
WORKDIR /build
COPY ./telegraf-serial ./telegraf-serial
RUN go build -o telegraf-serial cmd/main.go
FROM telegraf
WORKDIR /build
COPY --from=builder /build/telegraf-serial/telegraf-serial /bin/telegraf-serial
RUN mkdir -p /etc/telegraf/plugins
