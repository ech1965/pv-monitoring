FROM golang:alpine AS builder
WORKDIR /build
COPY ./telegraf-serial ./telegraf-serial
WORKDIR /build/telegraf-serial/cmd 
RUN go build -o telegraf-serial main.go
WORKDIR /build
FROM telegraf
WORKDIR /
COPY --from=builder /build/telegraf-serial/telegraf-serial /bin/telegraf-serial
RUN mkdir -p /etc/telegraf/plugins
