package database

import "sudojo/domain/lobby"

type Database interface {
	Close()
	InsertLobby(lobby *lobby.Lobby) error
	UpdateLobby(lobby *lobby.Lobby) error
	InsertPlayer(id string, player *lobby.Player) error
}
