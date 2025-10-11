package database

import (
	"sudojo/domain/lobby"

	"github.com/google/uuid"
)

type Database interface {
	Close()
	Lobby(id uuid.UUID, logger chan *lobby.Log) (*lobby.Lobby, error)
	InsertLobby(lobby *lobby.Lobby) error
	UpdateLobby(lobby *lobby.Lobby) error
	InsertPlayer(id string, player *lobby.Player) error
	InsertLogs(logs []*lobby.Log) error
}
