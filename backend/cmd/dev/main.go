package main

import (
	"fmt"
	"os"
	"sudojo/adapter/database"
	"sudojo/adapter/server"
	"sudojo/domain/lobby"
	"sudojo/service"
	"sudojo/service/conn"
	"sudojo/service/stats"
)

func main() {
	db := database.NewMock()
	logger := make(chan *lobby.Log, 1048576) // 2^20

	err := server.New(
		envOrPanic("PORT"),
		os.Getenv("ORIGIN"),
		[]service.Service{
			conn.New(logger, db, true),
			stats.New(logger, db),
		}).Listen()

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
