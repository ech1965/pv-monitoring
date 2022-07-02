FROM golang:alpine AS builder
WORKDIR /build
COPY ./telegraf-serial ./telegraf-serial
WORKDIR /build/telegraf-serial/cmd 
RUN go build -o telegraf-serial main.go
WORKDIR /build
FROM telegraf
WORKDIR /
RUN usermod -a -G dialout telegraf
RUN apt-get update; apt-get install musl
RUN mkdir -p /etc/telegraf/plugins 
COPY ./serial.conf /etc/telegraf/plugins/serial.conf
COPY --from=builder /build/telegraf-serial/cmd/telegraf-serial /bin/telegraf-serial
