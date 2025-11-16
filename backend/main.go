package main

import (
	"fmt"
	"os"
	"sudojo/adapter/server"
	"sudojo/service/tenant"
)

func main() {
	err := server.New(
		envOrPanic("PORT"),
		os.Getenv("ORIGIN"),
		tenant.New(),
	).Listen()

	if err != nil {
		panic(fmt.Errorf("failed starting server: %v", err))
	}
}

func envOrPanic(key string) string {
	val := os.Getenv(key)
	if len(val) < 1 {
		panic(fmt.Errorf("env variable %s missing", key))
	}
	return val
}
