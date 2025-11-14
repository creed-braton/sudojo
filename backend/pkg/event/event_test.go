package event

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type payload struct {
	Msg string `json:"msg"`
}

func (p *payload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func TestEvent(t *testing.T) {
	eventType, sender, trace, errorMsg, broadcast := "test-event", "test-sender", "test-trace", "test-error", true
	e := New(eventType, sender, trace, errorMsg, broadcast, &payload{})

	if e.Type() != eventType {
		t.Errorf("event type, want: %s, got: %s", eventType, e.Type())
	}
	if e.Sender() != sender {
		t.Errorf("sender, want: %s, got: %s", sender, e.Sender())
	}
	if e.Trace() != trace {
		t.Errorf("trace id, want: %s, got %s", trace, e.Trace())
	}
	if e.Error() != errorMsg {
		t.Errorf("error message, want: %s, got %s", errorMsg, e.Error())
	}
	if e.Broadcast() != broadcast {
		t.Errorf("broadcast flag, want: %t, got %t", broadcast, e.Broadcast())
	}
}

func TestEventChan(t *testing.T) {
	t.Run("send and receive event", func(t *testing.T) {
		channel := NewEventChan()
		defer channel.Close()

		want := "test-message"
		channel.Send(New("", "", "", "", false, &payload{Msg: want}))

		e, err := channel.Receive()
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

	t.Run("closed event channel", func(t *testing.T) {
		channel := NewEventChan()
		channel.Close()

		_, err := channel.Receive()
		if err != ErrClosedChan {
			t.Errorf("expected '%v' on receive, got: %v", ErrClosedChan, err)
		}
	})

	t.Run("full event channel", func(t *testing.T) {
		channel := NewEventChan()

		for range bufferSize {
			e := New("test", "sender", "receiver", "trace", false, nil)
			channel.Send(e)
		}
		channel.Send(New("test", "sender", "receiver", "trace", true, nil))

		for range bufferSize {
			e, err := channel.NonBlockRecv()
			if err != nil {
				t.Errorf("unexpected error '%v' on receive", err)
			}
			if e == nil {
				t.Error("expected event, got nil")
			}
		}
		e, err := channel.NonBlockRecv()
		if err != nil {
			t.Errorf("unexpected error '%v' on receive", err)
		}
		if e != nil {
			t.Error("expected nil, got event")
		}
	})
}

func TestEventBus(t *testing.T) {
	t.Run("broadcast delivery", func(t *testing.T) {
		bus := NewEventBus()
		defer bus.Close()

		num := 8
		numMsg := 32
		publisher := make(map[string]EventChan, num)
		subscriber := make(map[string]EventChan, num)
		for i := range num {
			pub := NewEventChan()
			sub := NewEventChan()
			id := fmt.Sprintf("worker-%d", i)
			publisher[id] = pub
			subscriber[id] = sub
			bus.Register(id, pub, sub)
		}

		for i := range numMsg {
			for id, pub := range publisher {
				trace := fmt.Sprintf("%d", i)
				e := New("test", id, trace, "", true, &payload{})
				pub.Send(e)
			}
			bus.Pump()
		}

		for id, sub := range subscriber {
			check := make(map[string]struct{}, num*numMsg)
			for range num * numMsg {
				e, err := sub.Receive()
				if err != nil {
					t.Errorf("unexpected error '%v' on receive", err)
				}
				msg := fmt.Sprintf("%s-%s", e.Sender(), e.Trace())
				if _, exist := check[msg]; exist {
					t.Errorf("duplicate message '%s' in '%s'", msg, id)
				}
				check[msg] = struct{}{}
			}

			for i := range num {
				for j := range numMsg {
					msg := fmt.Sprintf("worker-%d-%d", i, j)
					if _, exist := check[msg]; !exist {
						t.Errorf("missing message '%s' in '%s'", msg, id)
					}
				}
			}
		}
	})

	t.Run("targeted delivery", func(t *testing.T) {
		bus := NewEventBus()
		defer bus.Close()

		id := "sender"
		pub, sub, witness := NewEventChan(), NewEventChan(), NewEventChan()
		bus.Register(id, pub, sub)
		bus.Register("witness", NewEventChan(), witness)

		pub.Send(New("", id, "", "", false, &payload{}))
		bus.Pump()

		e, err := sub.Receive()
		if err != nil {
			t.Errorf("unexpected error '%v' on receive", err)
		}
		if e == nil {
			t.Error("expected event, got nil")
		}

		e, err = witness.NonBlockRecv()
		if err != nil {
			t.Errorf("unexpected error '%v' on receive", err)
		}
		if e != nil {
			t.Error("expected nil, got event")
		}
	})

	t.Run("close publisher channels", func(t *testing.T) {
		bus := NewEventBus()
		defer bus.Close()

		go func() {
			for {
				bus.Pump()
			}
		}()

		num := 8
		for i := range num {
			pub, sub := NewEventChan(), NewEventChan()
			bus.Register(fmt.Sprintf("worker-%d", i), pub, sub)

			go func() {
				for {
					if _, err := sub.Receive(); err != nil {
						return
					}
				}
			}()

			go func() {
				for range 1000 * i {
					pub.Send(New("", "", "", "", true, &payload{}))
				}
				pub.Close()
			}()
		}
	})

	t.Run("deregister channel", func(t *testing.T) {
		bus := NewEventBus()
		defer bus.Close()

		go func() {
			for {
				bus.Pump()
			}
		}()

		num := 8
		for i := range num {
			pub, sub := NewEventChan(), NewEventChan()
			bus.Register(fmt.Sprintf("worker-%d", i), pub, sub)

			go func() {
				for {
					if _, err := sub.Receive(); err != nil {
						return
					}
				}
			}()

			go func() {
				for {
					pub.Send(New("", "", "", "", true, &payload{}))
				}
			}()
		}

		time.Sleep(1 * time.Second)

		for i := range num {
			bus.Deregister(fmt.Sprintf("worker-%d", i))
		}
	})

	t.Run("re-register channel", func(t *testing.T) {
		bus := NewEventBus()
		defer bus.Close()

		go func() {
			for {
				bus.Pump()
			}
		}()

		num := 8
		for i := range num {
			pub, sub := NewEventChan(), NewEventChan()
			bus.Register(fmt.Sprintf("worker-%d", i), pub, sub)

			go func() {
				for {
					if _, err := sub.Receive(); err != nil {
						return
					}
				}
			}()

			go func() {
				for {
					pub.Send(New("", "", "", "", true, &payload{}))
				}
			}()
		}

		time.Sleep(1 * time.Second)

		for i := range num {
			pub, sub := NewEventChan(), NewEventChan()
			bus.Register(fmt.Sprintf("worker-%d", i), pub, sub)

			go func() {
				for {
					if _, err := sub.Receive(); err != nil {
						return
					}
				}
			}()

			go func() {
				for {
					pub.Send(New("", "", "", "", true, &payload{}))
				}
			}()
		}
	})
}

func BenchmarkEventBus(b *testing.B) {
	bus := NewEventBus()
	defer bus.Close()

	num := 16
	for i := range num - 1 {
		bus.Register(fmt.Sprintf("worker-%d", i), NewEventChan(), NewEventChan())
	}

	pub, sub := NewEventChan(), NewEventChan()
	bus.Register(fmt.Sprintf("worker-%d", num-1), pub, sub)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pub.Send(New("", "", "", "", true, &payload{}))
			bus.Pump()
			sub.Receive()
		}
	})
}
