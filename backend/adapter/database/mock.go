package database

import "sudojo/domain/lobby"

type mock struct{}

func NewMock() *mock {
	return &mock{}
}

func (db *mock) Close() {}

func (db *mock) InsertLobby(lobby *lobby.Lobby) error {
	return nil
}

func (db *mock) UpdateLobby(lobby *lobby.Lobby) error {
	return nil
}

func (db *mock) InsertPlayer(id string, player *lobby.Player) error {
	return nil
}

func (db *mock) InsertLogs(logs []*lobby.Log) error {
	return nil
}
