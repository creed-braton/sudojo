package manager

import (
	"fmt"
	"strconv"
	"sudojo/pkg/event"
	"sudojo/pkg/lobby"
	"sync"
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	t.Run("close manager", func(t *testing.T) {
		lobby := lobby.Open(false, 8)
		mng := New(lobby)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for {
				if err := mng.Pump(); err != nil {
					if err != event.ErrClosedBuffer {
						t.Errorf("expected '%v', got '%v'", event.ErrClosedBuffer, err)
					}
					wg.Done()
					return
				}
			}
		}()
		time.Sleep(time.Second)
		mng.Close()
		wg.Wait()
	})

	t.Run("join existing players", func(t *testing.T) {
		size := 8
		lobby := lobby.Open(false, size)
		mng := New(lobby)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for {
				if err := mng.Pump(); err != nil {
					wg.Done()
					return
				}
			}
		}()

		players := []Player{}
		for i := range size {
			token, err := mng.Create("")
			if err != nil {
				t.Errorf("unexpected error creating player '%v'", err)
			}
			p, err := mng.Join(token)
			if err != nil {
				t.Errorf("unexpected error joining player '%v'", err)
			}
			players = append(players, p)
			for _, p := range players {
				e, err := p.Receive()
				if err != nil {
					t.Errorf("unexpected error receiving event '%v'", err)
					continue
				}
				if e.Type() != event.JoinEvent {
					t.Errorf("expected event type '%s' got '%s'", event.JoinEvent, e.Type())
				}
				if e.Sender() != token {
					t.Errorf("expected sender '%s' got '%s'", token, e.Sender())
				}
				if len(e.Payload().Players()) != i+1 {
					t.Errorf("expected player list length '%d' got '%d'", i+1, len(e.Payload().Players()))
					continue
				}
				for _, p := range e.Payload().Players() {
					if !p.Active() {
						t.Error("unexpected inactive player")
					}
				}
			}
		}

		mng.Close()
		wg.Wait()
	})

	t.Run("join non-existing player", func(t *testing.T) {
		l := lobby.Open(false, 8)
		mng := New(l)
		defer mng.Close()
		_, err := mng.Join("test-token")
		if err == nil {
			t.Fatal("expected error joining player got nil")
		}
		if err != lobby.ErrPlayerNotFound {
			t.Errorf("expected error joining player '%v' got '%v'", err, lobby.ErrPlayerNotFound)
		}
	})

	t.Run("create player on full lobby", func(t *testing.T) {
		size := 8
		l := lobby.Open(false, size)
		mng := New(l)
		defer mng.Close()
		for range size {
			_, err := mng.Create("")
			if err != nil {
				t.Errorf("unexpected error creating player '%v'", err)
			}
		}
		_, err := mng.Create("")
		if err == nil {
			t.Error("expected error creating player got nil")
		}
		if err != lobby.ErrLobbyFull {
			t.Errorf("expected error creating player '%v' got '%v'", err, lobby.ErrLobbyFull)
		}
	})

	t.Run("leave player", func(t *testing.T) {
		size := 8
		lobby := lobby.Open(false, size)
		mng := New(lobby)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			for {
				if err := mng.Pump(); err != nil {
					wg.Done()
					return
				}
			}
		}()

		players := []Player{}
		tokens := []string{}
		for i := range size {
			token, err := mng.Create(fmt.Sprintf("%d", i))
			if err != nil {
				t.Errorf("unexpected error creating player '%v'", err)
			}
			p, err := mng.Join(token)
			if err != nil {
				t.Errorf("unexpected error joining player '%v'", err)
			}
			players = append(players, p)
			tokens = append(tokens, token)
			for _, p := range players {
				p.Receive()
			}
		}

		for i, token := range tokens {
			mng.leave(token)
			for _, p := range players {
				e, err := p.Receive()
				if err != nil {
					t.Errorf("unexpected error recieving event '%v'", err)
					continue
				}
				if e.Type() != event.LeaveEvent {
					t.Errorf("expected event type '%s' got '%s'", event.LeaveEvent, e.Type())
				}
				if e.Sender() != token {
					t.Errorf("expected sender '%s' got '%s'", token, e.Sender())
				}
				for _, p := range e.Payload().Players() {
					name, err := strconv.Atoi(p.Name())
					if err != nil {
						t.Errorf("unexpected error parsing player name '%v'", err)
					}
					if name <= i && p.Active() {
						t.Errorf("expected player '%d' to be inactive ", name)
					} else if name > i && !p.Active() {
						t.Errorf("expected player '%d' to be active ", name)
					}
				}
			}
		}

		mng.Close()
		wg.Wait()
	})
}
