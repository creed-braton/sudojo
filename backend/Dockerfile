FROM golang:1.25.1-alpine3.22 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd

COPY domain ./domain

COPY adapter ./adapter

COPY service ./service

RUN go build -o bin cmd/app/main.go

FROM alpine:3.22

COPY --from=build /app/bin /main

ENV PORT=8080

ENTRYPOINT [ "/main" ]
