FROM golang:1.25.1 AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG CMD
RUN go build -o ./bin ./cmd/${CMD}/main.go

FROM scratch

COPY --from=build /build/bin /main

ENTRYPOINT [ "/main" ]
