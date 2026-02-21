package database

import (
	"sudojo/pkg/game"
	"sudojo/pkg/lobby"
)

type mock struct{}

var _ Database = &mock{}

func NewMock() *mock {
	return &mock{}
}

func (db *mock) Close() {}

func (db *mock) Lobby(id string) (lobby.Lobby, error) {
	return nil, nil
}

func (db *mock) InsertLobby(lobby lobby.Lobby) error {
	return nil
}

func (db *mock) UpdateLobby(lobby lobby.Lobby) error {
	return nil
}

func (db *mock) InsertPlayer(id, token, name string) error {
	return nil
}

func (db *mock) SampleGame(difficulty string) (game.Game, error) {
	return nil, nil
}
