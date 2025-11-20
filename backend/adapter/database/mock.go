package database

import "sudojo/pkg/lobby"

type mockDatabase struct{}

var _ Database = &mockDatabase{}

func NewMock() *mockDatabase {
	return &mockDatabase{}
}

func (db *mockDatabase) Close() {}

func (db *mockDatabase) Lobby(id string) (lobby.Lobby, error) {
	return nil, nil
}

func (db *mockDatabase) InsertLobby(lobby lobby.Lobby) error {
	return nil
}

func (db *mockDatabase) UpdateLobby(lobby lobby.Lobby) error {
	return nil
}

func (db *mockDatabase) InsertPlayer(lobbyId, token, name string) error {
	return nil
}
