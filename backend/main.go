package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sudojo/adapter/database"
	"sudojo/adapter/server"
	"sudojo/domain/lobby"
	"sudojo/service"
	"sudojo/service/conn"
	"sudojo/service/stats"
)

func main() {
	logger := make(chan *lobby.Log, 1048576) // 2^20
	var db database.Database
	var err error

	if dev() {
		db = database.NewMock()
		log.Println("WARNING: starting server without a db connection")
	} else {
		db, err = database.New(
			envOrPanic("DB_HOST"),
			envOrPanic("DB_PORT"),
			envOrPanic("DB_NAME"),
			envOrPanic("DB_USER"),
			envOrPanic("DB_PASS"),
		)
		if err != nil {
			panic(fmt.Errorf("failed establishing db connection: %v", err))
		} else {
			log.Println("INFO: successfully connected to db")
		}
	}

	err = server.New(
		envOrPanic("PORT"),
		os.Getenv("ORIGIN"),
		[]service.Service{
			conn.New(logger, db),
			stats.New(logger, db),
		}).Listen()

	if err != nil {
		panic(fmt.Errorf("failed starting server: %v", err))
	}
}

func dev() bool {
	return strings.ToLower(os.Getenv("DEV")) == "true"
}

func envOrPanic(key string) string {
	val := os.Getenv(key)
	if len(val) < 1 {
		panic(fmt.Errorf("env variable %s missing", key))
	}
	return val
}
