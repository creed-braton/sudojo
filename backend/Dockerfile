FROM golang:1.25.1-alpine3.22 AS build

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./main.go
COPY pkg ./pkg
COPY adapter ./adapter
COPY service ./service

RUN go build -o ./bin ./main.go

FROM alpine:3.22

COPY --from=build /build/bin /main

ENV PORT=8080

ENTRYPOINT [ "/main" ]
