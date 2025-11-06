package event

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

type payload struct {
	Id  int    `json:"id"`
	Msg string `json:"msg"`
}

func (p *payload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func TestEvent(t *testing.T) {
	eventType, sender, receiver, trace, errorMsg := "test-event", "test-sender", "test-receiver", "test-trace", "test-error"
	e := New(eventType, sender, receiver, trace, errorMsg, &payload{})

	if e.Type() != eventType {
		t.Errorf("want: %s, got %s", eventType, e.Type())
	}
	if e.Sender() != sender {
		t.Errorf("want: %s, got %s", sender, e.Sender())
	}
	if e.Receiver() != receiver {
		t.Errorf("want: %s, got %s", receiver, e.Receiver())
	}
	if e.Trace() != trace {
		t.Errorf("want: %s, got %s", trace, e.Trace())
	}
	if e.Error() != errorMsg {
		t.Errorf("want: %s, got %s", errorMsg, e.Error())
	}
}

func TestEventBus(t *testing.T) {
	t.Run("send and receive event", func(t *testing.T) {
		bus := NewEventBus()
		defer bus.Close()

		want := "test-message"
		var e Event = New("", "", "", "", "", &payload{Id: 0, Msg: want})
		if err := bus.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}

		e, err := bus.Receive()
		if err != nil {
			t.Fatalf("unexpected error receiving event: %v", err)
		}

		p, ok := e.Payload().(*payload)
		if !ok {
			t.Fatalf("unexpected payload type: %T", e.Payload())
		}

		if p.Msg != want {
			t.Errorf("want: %s, got: %s", want, p.Msg)
		}
	})

	t.Run("closed event bus", func(t *testing.T) {
		bus := NewEventBus()
		bus.Close()

		err := bus.Send(New("closed-test", "sender", "", "", "", &payload{Msg: "test"}))
		if err != ErrClosedBus {
			t.Errorf("expected '%v' on send, got: %v", ErrClosedBus, err)
		}

		_, err = bus.Receive()
		if err != ErrClosedBus {
			t.Errorf("expected '%v' on receive, got: %v", ErrClosedBus, err)
		}
	})

	t.Run("full event bus", func(t *testing.T) {
		bus := NewEventBus()

		for range eventBusSize {
			e := New("test", "sender", "receiver", "trace", "", nil)
			if err := bus.Send(e); err != nil {
				t.Fatalf("unexpected error sending event: %v", err)
			}
		}

		err := bus.Send(New("test", "sender", "receiver", "trace", "", nil))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrFullBus {
			t.Errorf("expected '%v' on receive, got: %v", ErrFullBus, err)
		}
	})

	t.Run("concurrent senders", func(t *testing.T) {
		const (
			numSenders = 8
			numEvents  = 32
			total      = numSenders * numEvents
		)

		bus := NewEventBus()
		defer bus.Close()

		var wg sync.WaitGroup

		// Start concurrent senders
		for senderId := range numSenders {
			wg.Add(1)
			go func(senderId int) {
				defer wg.Done()
				for eventId := range numEvents {
					msg := fmt.Sprintf("%d-%d", senderId, eventId)
					e := New("test-type", fmt.Sprintf("sender-%d", senderId), "", msg, "", &payload{Msg: msg})
					if err := bus.Send(e); err != nil {
						t.Errorf("send error from sender %d: %v", senderId, err)
						return
					}
				}
			}(senderId)
		}

		received := make(map[string]bool)
		for i := 0; i < total; i++ {
			e, err := bus.Receive()
			if err != nil {
				t.Errorf("receive error: %v", err)
				return
			}

			p, ok := e.Payload().(*payload)
			if !ok {
				t.Errorf("unexpected payload type: %T", e.Payload())
				continue
			}

			if received[p.Msg] {
				t.Errorf("duplicate message received: %s", p.Msg)
			}
			received[p.Msg] = true
		}

		wg.Wait()

		for senderId := range numSenders {
			for eventId := range numEvents {
				msg := fmt.Sprintf("%d-%d", senderId, eventId)
				if !received[msg] {
					t.Errorf("missing message: %s", msg)
				}
			}
		}
	})

	t.Run("concurrent senders on closed event bus", func(t *testing.T) {
		for range 100 {
			bus := NewEventBus()
			done := make(chan struct{})

			for range 1000 {
				go func() {
					for {
						select {
						case <-done:
							return
						default:
							bus.Send(New("test", "sender", "receiver", "trace", "", nil))
						}
					}
				}()
			}

			bus.Close()
			time.Sleep(1 * time.Second)
			close(done)
		}
	})
}
