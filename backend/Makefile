dev: build dev-run clean

build:
	@go build -o ./bin ./main.go

dev-run:
	@DEV=true ORIGIN=http://localhost:5173 PORT="8080" ./bin

test:
	@go test -v ./...

clean:
	@rm ./bin
