all: build run clean

build:
	@go build -o ./bin ./main.go

run:
	@ORIGIN=http://localhost:5173 PORT="8080" DB_HOST="localhost" DB_PORT="5432" DB_NAME="postgres" DB_USER="postgres" DB_PASS="password" ./bin

test:
	@go test -v ./...

clean:
	@rm ./bin
