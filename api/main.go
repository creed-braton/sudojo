package main

import (
	"fmt"
	"os"
	"sudojo/adapter/database"
	"sudojo/adapter/metrics"
	"sudojo/adapter/server"
	"sudojo/service/tenant"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	m := metrics.New(prometheus.DefaultRegisterer)

	db, err := database.New(
		envOrPanic("DB_HOST"),
		envOrPanic("DB_PORT"),
		envOrPanic("DB_NAME"),
		envOrPanic("DB_USER"),
		envOrPanic("DB_PASS"),
	)
	if err != nil {
		panic(err)
	}
	err = server.New(
		envOrPanic("PORT"),
		os.Getenv("ORIGIN"),
		tenant.New(db, m),
	).Listen()

	if err != nil {
		panic(fmt.Errorf("failed starting server: %v", err))
	}
}

func envOrPanic(key string) string {
	val := os.Getenv(key)
	if len(val) < 1 {
		panic(fmt.Errorf("env variable '%s' missing", key))
	}
	return val
}
