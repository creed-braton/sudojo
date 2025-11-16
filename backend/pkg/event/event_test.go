package event

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestEvent(t *testing.T) {
	eventType, sender, trace, errorMsg := "test-event", "test-sender", "test-trace", "test-error"
	e := New(eventType, sender, trace, errorMsg, &payload{})

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
}

func TestEventBus(t *testing.T) {
	t.Run("send and receive event", func(t *testing.T) {
		bus := NewEventBus()
		defer bus.Close()

		want := "test-message"
		var e Event = New("", "", want, "", NewPayload())
		if err := bus.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}

		e, err := bus.Receive()
		if err != nil {
			t.Fatalf("unexpected error receiving event: %v", err)
		}

		if e.Trace() != want {
			t.Errorf("expected: '%s', got: '%s'", want, e.Trace())
		}
	})

	t.Run("closed event bus", func(t *testing.T) {
		bus := NewEventBus()
		bus.Close()

		err := bus.Send(New("closed-test", "sender", "", "", NewPayload()))
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
			e := New("test", "sender", "receiver", "trace", nil)
			if err := bus.Send(e); err != nil {
				t.Fatalf("unexpected error sending event: %v", err)
			}
		}

		err := bus.Send(New("test", "sender", "receiver", "trace", nil))
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
					e := New("test-type", fmt.Sprintf("sender-%d", senderId), msg, "", NewPayload())
					if err := bus.Send(e); err != nil {
						t.Errorf("send error from sender %d: %v", senderId, err)
						return
					}
				}
			}(senderId)
		}

		received := make(map[string]bool)
		for range total {
			e, err := bus.Receive()
			if err != nil {
				t.Errorf("receive error: %v", err)
				return
			}

			if received[e.Trace()] {
				t.Errorf("duplicate message received: %s", e.Trace())
			}
			received[e.Trace()] = true
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
							bus.Send(New("test", "sender", "receiver", "trace", nil))
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

func TestFanout(t *testing.T) {
	t.Run("broadcast delivery", func(t *testing.T) {
		src := NewEventBus()
		defer src.Close()
		fanout := NewFanout(src)

		numRecv := 8
		receiver := make(map[string]EventBus)
		for i := range numRecv {
			bus := NewEventBus()
			id := fmt.Sprintf("receiver-%d", i)
			receiver[id] = bus
			fanout.Register(id, bus)
		}

		e := New("test", fmt.Sprintf("receiver-%d", 0), "", "", nil)
		if err := src.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}
		if err := fanout.Pump(); err != nil {
			t.Fatalf("unexpected error pumping fanout: %v", err)
		}

		for k, v := range receiver {
			e, err := v.Receive()
			if err != nil {
				t.Fatalf("unexpected error receiving event: %v", err)
			}
			if e == nil {
				t.Fatalf("received nil instead of event for receiver '%s'", k)
			}
		}
	})

	t.Run("deregister target event bus", func(t *testing.T) {
		src := NewEventBus()
		fanout := NewFanout(src)

		numRecv := 8
		receiver := make(map[string]EventBus)
		for i := range numRecv {
			bus := NewEventBus()
			id := fmt.Sprintf("receiver-%d", i)
			receiver[id] = bus
			fanout.Register(id, bus)
		}

		id := fmt.Sprintf("receiver-%d", 5)
		bus := receiver[id]
		if bus == nil {
			t.Fatalf("%s event bus not stored in receiver map", id)
		}
		fanout.Deregister(id)

		e := New("test", fmt.Sprintf("receiver-%d", 0), "", "", nil)
		if err := src.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}
		if err := fanout.Pump(); err != nil {
			t.Fatalf("unexpected error pumping fanout: %v", err)
		}
		src.Close()

		go func() {
			e, err := bus.Receive()
			t.Error("unexpected resolved block on deregistered event bus receive")
			if err != nil {
				t.Errorf("unexpected error receiving in deregistered event bus: %v", err)
			}
			if e != nil {
				t.Error("received unexpected event in deregisterd event bus")
			}
		}()

		time.Sleep(3 * time.Second)
	})

	t.Run("re-register existing id", func(t *testing.T) {
		src := NewEventBus()
		defer src.Close()
		fanout := NewFanout(src)

		id := "test-id"
		old := NewEventBus()
		fanout.Register(id, old)
		bus := NewEventBus()
		fanout.Register(id, bus)
		if err := src.Send(New("", "", "", "", nil)); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}
		if err := fanout.Pump(); err != nil {
			t.Fatalf("unexpected error pumping fanout: %v", err)
		}

		if _, err := old.Receive(); err == nil {
			t.Errorf("expected '%v', got nil", ErrClosedBus)
		} else if err != ErrClosedBus {
			t.Errorf("expected '%v', got: %v", ErrClosedBus, err)
		}

		e, err := bus.Receive()
		if err != nil {
			t.Errorf("unexpected error receiving event: %v", err)
		}
		if e == nil {
			t.Error("expected event got nil")
		}
	})

	t.Run("closed target event bus", func(t *testing.T) {
		src := NewEventBus()
		defer src.Close()
		fanout := NewFanout(src)

		numRecv := 8
		receiver := make(map[string]EventBus)
		for i := range numRecv {
			bus := NewEventBus()
			id := fmt.Sprintf("receiver-%d", i)
			receiver[id] = bus
			fanout.Register(id, bus)
		}

		id := fmt.Sprintf("receiver-%d", 6)
		bus := receiver[id]
		if bus == nil {
			t.Fatalf("%s event bus not stored in receiver map", id)
		}
		bus.Close()

		e := New("test", fmt.Sprintf("receiver-%d", 0), "", "", nil)
		if err := src.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}
		if err := fanout.Pump(); err != nil {
			t.Fatalf("unexpected error pumping fanout: %v", err)
		}

		if fanout.routes[id] != nil {
			t.Errorf("event bus '%s' still registered", id)
		}
	})

	t.Run("closed source event bus", func(t *testing.T) {
		src := NewEventBus()
		fanout := NewFanout(src)

		receiver := []EventBus{}
		for i := range 8 {
			bus := NewEventBus()
			fanout.Register(fmt.Sprintf("receiver-%d", i), bus)
			receiver = append(receiver, bus)
		}

		src.Close()
		if err := fanout.Pump(); err != ErrClosedBus {
			t.Fatalf("unexpected error pumping fanout: %v, want: %v", err, ErrClosedBus)
		}

		for _, bus := range receiver {
			if _, err := bus.Receive(); err != ErrClosedBus {
				t.Errorf("want: %v, got: %v", ErrClosedBus, err)
			}
		}
	})

	t.Run("under load", func(t *testing.T) {
		src := NewEventBus()
		fanout := NewFanout(src)

		// make sure that numMsg * numWorker does not exceed eventBusSize
		numMsg := 16
		numWorker := 16
		workers := make(map[string]EventBus)
		for i := range numWorker {
			bus := NewEventBus()
			id := fmt.Sprintf("worker-%d", i)
			workers[id] = bus
			fanout.Register(id, bus)
		}

		wg := sync.WaitGroup{}

		// polling go routine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range numWorker * numMsg {
				if err := fanout.Pump(); err != nil {
					t.Errorf("unexpected error pumping in %d: %v", i, err)
				}
			}
		}()

		// receiver go routines
		for k, v := range workers {
			wg.Add(1)
			go func(id string, bus EventBus) {
				defer wg.Done()
				check := make(map[string]struct{})
				for i := range numWorker * numMsg {
					e, err := bus.Receive()
					if err != nil {
						t.Errorf("unexpected error receiving in %s iteration %d: %v", id, i, err)
						continue
					}

					if e == nil {
						t.Errorf("nil event in worker %s", id)
						continue
					}

					eventId := fmt.Sprintf("%s-%s", e.Sender(), e.Trace())
					if _, exist := check[eventId]; exist {
						t.Errorf("duplicate message '%s' in %s", eventId, id)
					} else {
						check[eventId] = struct{}{}
					}
				}

				for i := range numWorker {
					for j := range numMsg {
						eventId := fmt.Sprintf("worker-%d-%d", i, j)
						if _, exist := check[eventId]; !exist {
							t.Errorf("missing message '%s' in %s", eventId, id)
						}
					}
				}
			}(k, v)
		}

		// sender go routines
		for k := range workers {
			wg.Add(1)
			go func(id string, bus EventBus) {
				defer wg.Done()
				for i := range numMsg {
					e := New("test", id, strconv.Itoa(i), "", nil)
					if err := bus.Send(e); err != nil {
						t.Errorf("unexpected error sending in %s iteration %d: %v", id, i, err)
					}
				}
			}(k, src)
		}

		wg.Wait()
	})
}

func BenchmarkFanout(b *testing.B) {
	src := NewEventBus()
	defer src.Close()
	fanout := NewFanout(src)

	numWorker := 16
	for i := range numWorker - 1 {
		fanout.Register(fmt.Sprintf("worker-%d", i), NewEventBus())
	}

	target := NewEventBus()
	fanout.Register(fmt.Sprintf("worker-%d", numWorker-1), target)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			src.Send(New("", "", "", "", nil))
			fanout.Pump()
			target.Receive()
		}
	})
}
