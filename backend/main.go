package main

import (
	"sudojo/adapter/server"
	"sudojo/service"
	"sudojo/service/lobby"
)

func main() {
	err := server.New(
		":8080",
		[]service.Service{lobby.New()},
	).Listen()

	if err != nil {
		panic(err)
	}
}
