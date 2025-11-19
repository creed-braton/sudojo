package database

import "sudojo/pkg/lobby"

type mockDatabase struct {
	lobby        func(id string) (lobby.Lobby, error)
	insertLobby  func(lobby lobby.Lobby) error
	updateLobby  func(lobby lobby.Lobby) error
	insertPlayer func(lobbyId, token, name string) error
}

var _ Database = &mockDatabase{}

func NewMockDatabase(
	lobby func(id string) (lobby.Lobby, error),
	insertLobby func(lobby lobby.Lobby) error,
	updateLobby func(lobby lobby.Lobby) error,
	insertPlayer func(lobbyId, token, name string) error,
) *mockDatabase {
	return &mockDatabase{
		lobby:        lobby,
		insertLobby:  insertLobby,
		updateLobby:  updateLobby,
		insertPlayer: insertPlayer,
	}
}

func (db *mockDatabase) Close() {}

func (db *mockDatabase) Lobby(id string) (lobby.Lobby, error) {
	return db.lobby(id)
}

func (db *mockDatabase) InsertLobby(lobby lobby.Lobby) error {
	return db.insertLobby(lobby)
}

func (db *mockDatabase) UpdateLobby(lobby lobby.Lobby) error {
	return db.updateLobby(lobby)
}

func (db *mockDatabase) InsertPlayer(lobbyId, token, name string) error {
	return db.insertPlayer(lobbyId, token, name)
}
