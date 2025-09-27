package main

import (
	"sudojo/adapter/server"
	"sudojo/service"
	"sudojo/service/conn"
)

func main() {
	err := server.New(
		":8080",
		[]service.Service{conn.New()},
	).Listen()

	if err != nil {
		panic(err)
	}
}
