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
	e := New().
		SetType(eventType).
		SetSender(sender).
		SetTrace(trace).
		SetError(errorMsg).
		SetPayload(NewPayload())

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
		buffer := NewBuffer()
		defer buffer.Close()

		want := "test-message"
		var e Event = New().SetTrace(want).SetPayload(NewPayload())
		if err := buffer.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}

		e, err := buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected error receiving event: %v", err)
		}

		if e.Trace() != want {
			t.Errorf("expected: '%s', got: '%s'", want, e.Trace())
		}
	})

	t.Run("closed event buffer", func(t *testing.T) {
		buffer := NewBuffer()
		buffer.Close()

		err := buffer.Send(New().SetType("closed-test").SetSender("sender").SetPayload(NewPayload()))
		if err != ErrClosedBuffer {
			t.Errorf("expected '%v' on send, got: %v", ErrClosedBuffer, err)
		}

		_, err = buffer.Receive()
		if err != ErrClosedBuffer {
			t.Errorf("expected '%v' on receive, got: %v", ErrClosedBuffer, err)
		}
	})

	t.Run("full event buffer", func(t *testing.T) {
		buffer := NewBuffer()

		for range bufferSize {
			e := New().SetType("test").SetSender("sender").SetTrace("trace")
			if err := buffer.Send(e); err != nil {
				t.Fatalf("unexpected error sending event: %v", err)
			}
		}

		err := buffer.Send(New().SetType("test").SetSender("sender").SetTrace("trace"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrFullBuffer {
			t.Errorf("expected '%v' on receive, got: %v", ErrFullBuffer, err)
		}
	})

	t.Run("concurrent senders", func(t *testing.T) {
		const (
			numSenders = 8
			numEvents  = 32
			total      = numSenders * numEvents
		)

		buffer := NewBuffer()
		defer buffer.Close()

		var wg sync.WaitGroup

		// Start concurrent senders
		for senderId := range numSenders {
			wg.Add(1)
			go func(senderId int) {
				defer wg.Done()
				for eventId := range numEvents {
					msg := fmt.Sprintf("%d-%d", senderId, eventId)
					e := New().
						SetType("test-type").
						SetSender(fmt.Sprintf("sender-%d", senderId)).
						SetTrace(msg).
						SetPayload(NewPayload())
					if err := buffer.Send(e); err != nil {
						t.Errorf("send error from sender %d: %v", senderId, err)
						return
					}
				}
			}(senderId)
		}

		received := make(map[string]bool)
		for range total {
			e, err := buffer.Receive()
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

	t.Run("concurrent senders on closed event buffer", func(t *testing.T) {
		for range 100 {
			buffer := NewBuffer()
			done := make(chan struct{})

			for range 1000 {
				go func() {
					for {
						select {
						case <-done:
							return
						default:
							buffer.Send(New().SetType("test").SetSender("sender").SetTrace("trace"))
						}
					}
				}()
			}

			buffer.Close()
			time.Sleep(1 * time.Second)
			close(done)
		}
	})
}

func TestFanout(t *testing.T) {
	t.Run("broadcast delivery", func(t *testing.T) {
		src := NewBuffer()
		defer src.Close()
		fanout := NewFanout(src)

		numRecv := 8
		receiver := make(map[string]Buffer)
		for i := range numRecv {
			buffer := NewBuffer()
			id := fmt.Sprintf("receiver-%d", i)
			receiver[id] = buffer
			fanout.Register(id, buffer)
		}

		e := New().SetType("test").SetSender(fmt.Sprintf("receiver-%d", 0))
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

	t.Run("deregister target event buffer", func(t *testing.T) {
		src := NewBuffer()
		fanout := NewFanout(src)

		numRecv := 8
		receiver := make(map[string]Buffer)
		for i := range numRecv {
			buffer := NewBuffer()
			id := fmt.Sprintf("receiver-%d", i)
			receiver[id] = buffer
			fanout.Register(id, buffer)
		}

		id := fmt.Sprintf("receiver-%d", 5)
		buffer := receiver[id]
		if buffer == nil {
			t.Fatalf("%s event buffer not stored in receiver map", id)
		}
		fanout.Deregister(id)

		e := New().SetType("test").SetSender(fmt.Sprintf("receiver-%d", 0))
		if err := src.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}
		if err := fanout.Pump(); err != nil {
			t.Fatalf("unexpected error pumping fanout: %v", err)
		}
		src.Close()

		go func() {
			e, err := buffer.Receive()
			t.Error("unexpected resolved block on deregistered event buffer receive")
			if err != nil {
				t.Errorf("unexpected error receiving in deregistered event buffer: %v", err)
			}
			if e != nil {
				t.Error("received unexpected event in deregisterd event buffer")
			}
		}()

		time.Sleep(3 * time.Second)
	})

	t.Run("re-register existing id", func(t *testing.T) {
		src := NewBuffer()
		defer src.Close()
		fanout := NewFanout(src)

		id := "test-id"
		old := NewBuffer()
		fanout.Register(id, old)
		buffer := NewBuffer()
		fanout.Register(id, buffer)
		if err := src.Send(New()); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}
		if err := fanout.Pump(); err != nil {
			t.Fatalf("unexpected error pumping fanout: %v", err)
		}

		if _, err := old.Receive(); err == nil {
			t.Errorf("expected '%v', got nil", ErrClosedBuffer)
		} else if err != ErrClosedBuffer {
			t.Errorf("expected '%v', got: %v", ErrClosedBuffer, err)
		}

		e, err := buffer.Receive()
		if err != nil {
			t.Errorf("unexpected error receiving event: %v", err)
		}
		if e == nil {
			t.Error("expected event got nil")
		}
	})

	t.Run("closed target event buffer", func(t *testing.T) {
		src := NewBuffer()
		defer src.Close()
		fanout := NewFanout(src)

		numRecv := 8
		receiver := make(map[string]Buffer)
		for i := range numRecv {
			buffer := NewBuffer()
			id := fmt.Sprintf("receiver-%d", i)
			receiver[id] = buffer
			fanout.Register(id, buffer)
		}

		id := fmt.Sprintf("receiver-%d", 6)
		buffer := receiver[id]
		if buffer == nil {
			t.Fatalf("%s event buffer not stored in receiver map", id)
		}
		buffer.Close()

		e := New().SetType("test").SetSender(fmt.Sprintf("receiver-%d", 0))
		if err := src.Send(e); err != nil {
			t.Fatalf("unexpected error sending event: %v", err)
		}
		if err := fanout.Pump(); err != nil {
			t.Fatalf("unexpected error pumping fanout: %v", err)
		}

		if fanout.routes[id] != nil {
			t.Errorf("event buffer '%s' still registered", id)
		}
	})

	t.Run("closed source event buffer", func(t *testing.T) {
		src := NewBuffer()
		fanout := NewFanout(src)

		receiver := []Buffer{}
		for i := range 8 {
			buffer := NewBuffer()
			fanout.Register(fmt.Sprintf("receiver-%d", i), buffer)
			receiver = append(receiver, buffer)
		}

		src.Close()
		if err := fanout.Pump(); err != ErrClosedBuffer {
			t.Fatalf("unexpected error pumping fanout: %v, want: %v", err, ErrClosedBuffer)
		}

		for _, buffer := range receiver {
			if _, err := buffer.Receive(); err != ErrClosedBuffer {
				t.Errorf("want: %v, got: %v", ErrClosedBuffer, err)
			}
		}
	})

	t.Run("under load", func(t *testing.T) {
		src := NewBuffer()
		fanout := NewFanout(src)

		// make sure that numMsg * numWorker does not exceed bufferSize
		numMsg := 16
		numWorker := 16
		workers := make(map[string]Buffer)
		for i := range numWorker {
			buffer := NewBuffer()
			id := fmt.Sprintf("worker-%d", i)
			workers[id] = buffer
			fanout.Register(id, buffer)
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
			go func(id string, buffer Buffer) {
				defer wg.Done()
				check := make(map[string]struct{})
				for i := range numWorker * numMsg {
					e, err := buffer.Receive()
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
			go func(id string, buffer Buffer) {
				defer wg.Done()
				for i := range numMsg {
					e := New().SetType("test").SetSender(id).SetTrace(strconv.Itoa(i))
					if err := buffer.Send(e); err != nil {
						t.Errorf("unexpected error sending in %s iteration %d: %v", id, i, err)
					}
				}
			}(k, src)
		}

		wg.Wait()
	})
}

func BenchmarkFanout(b *testing.B) {
	src := NewBuffer()
	defer src.Close()
	fanout := NewFanout(src)

	numWorker := 16
	for i := range numWorker - 1 {
		fanout.Register(fmt.Sprintf("worker-%d", i), NewBuffer())
	}

	target := NewBuffer()
	fanout.Register(fmt.Sprintf("worker-%d", numWorker-1), target)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			src.Send(New())
			fanout.Pump()
			target.Receive()
		}
	})
}
