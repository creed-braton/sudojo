dev: dev-build dev-run clean

dev-build:
	@go build -o ./bin ./cmd/dev

dev-run:
	@ORIGIN=http://localhost:5173 PORT="8080" ./bin

test:
	@go test -v ./...

clean:
	@rm ./bin
