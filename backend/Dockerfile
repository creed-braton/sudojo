FROM golang:1.25.1-alpine3.21 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./main.go

COPY domain ./domain

COPY adapter ./adapter

COPY service ./service

RUN go build -o bin main.go

FROM alpine:3.21

COPY --from=build /app/bin /main

ENTRYPOINT [ "/main" ]
