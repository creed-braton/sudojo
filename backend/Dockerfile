FROM golang:1.25.1 AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg
COPY adapter ./adapter
COPY service ./service
COPY main.go ./main.go

RUN go build -o ./bin

FROM scratch

COPY --from=build /build/bin /main

ENTRYPOINT [ "/main" ]
