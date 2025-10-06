package main

import (
	"sudojo/adapter/server"
	"sudojo/domain/lobby"
	"sudojo/service"
	"sudojo/service/conn"
	"sudojo/service/stats"
)

func main() {
	complete := make(chan *lobby.Lobby, 1024)
	err := server.New(
		":8080", []service.Service{
			conn.New(complete),
			stats.New(complete),
		}).Listen()

	if err != nil {
		panic(err)
	}
}
