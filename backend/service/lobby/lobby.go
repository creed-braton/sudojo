package lobby

import (
	"sudojo/adapter/socket"
	"sudojo/pkg/lobby"

	"github.com/gorilla/websocket"
)

type Service interface {
}

type service struct {
	lobby lobby.Lobby
}

func New() *service {
	return &service{}
}

func (s *service) receive() {

}

func (s *service) Join(token, name string, conn *websocket.Conn) error {
	client := socket.NewClient(conn)
	player, err := s.lobby.Join(token, name)
	if err != nil {
		client.Close()
		return err
	}

	go client.ReadPump()
	go client.WritePump()
}
