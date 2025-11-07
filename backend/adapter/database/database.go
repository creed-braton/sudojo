package database

import (
	"sudojo/internal/domain/lobby"

	"github.com/google/uuid"
)

type Database interface {
	Close()
	Lobby(id uuid.UUID) (*lobby.Lobby, error)
	InsertLobby(lobby *lobby.Lobby) error
	UpdateLobby(lobby *lobby.Lobby) error
	InsertPlayer(id string, player *lobby.Player) error
}
