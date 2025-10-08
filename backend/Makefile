all: build run

test:
	@go test -v ./...

build:
	@go build -tags dev -o ./bin ./cmd/dev

run:
	@ORIGIN=http://localhost:5173 PORT="8080" ./bin

clean:
	@rm ./bin
