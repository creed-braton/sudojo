package event

import (
	"sync"
	"testing"
)

func TestNewHub(t *testing.T) {
	t.Parallel()

	hub := NewHub()

	buffer := NewBuffer(1)
	defer buffer.Close(0)

	if err := hub.Register("test", buffer); err != nil {
		t.Errorf("expected nil error on fresh hub, got: '%v'", err)
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("single buffer", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		buffer := NewBuffer(8)
		defer hub.Close(0)

		if err := hub.Register("player-1", buffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		event := New(JoinEvent, int64(42), "trace")
		hub.Broadcast(event)

		received, err := buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}

		if received.Trace() != event.Trace() {
			t.Errorf("expected trace '%s', got: '%s'", event.Trace(), received.Trace())
		}
	})

	t.Run("multiple buffers", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		buffers := make([]*buffer, 5)
		for i := range buffers {
			buffers[i] = NewBuffer(8)
			if err := hub.Register(string(rune('a'+i)), buffers[i]); err != nil {
				t.Fatalf("unexpected register error for buffer %d: '%v'", i, err)
			}
		}

		event := New(StateEvent, int64(42), "broadcast-test")
		hub.Broadcast(event)

		for i, buf := range buffers {
			received, err := buf.Receive()
			if err != nil {
				t.Fatalf("buffer %d: unexpected receive error: '%v'", i, err)
			}
			if received.Trace() != event.Trace() {
				t.Errorf("buffer %d: expected trace '%s', got: '%s'", i, event.Trace(), received.Trace())
			}
		}
	})

	t.Run("takeover", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		oldBuffer := NewBuffer(8)
		newBuffer := NewBuffer(8)

		if err := hub.Register("player-1", oldBuffer); err != nil {
			t.Fatalf("unexpected register error for old buffer: '%v'", err)
		}

		if err := hub.Register("player-1", newBuffer); err != nil {
			t.Fatalf("unexpected register error for new buffer: '%v'", err)
		}

		// Old buffer should be closed with TakeoverReason
		err := oldBuffer.Send(New(PingEvent, int64(42), ""))
		if err == nil {
			t.Fatal("expected error on old buffer, got nil")
		}
		if closeErr, ok := err.(*BufferClosedError); !ok {
			t.Errorf("expected BufferClosedError, got: '%v'", err)
		} else {
			if closeErr.Reason() != TakeoverReason {
				t.Errorf("expected TakeoverReason (%d), got: %d", TakeoverReason, closeErr.Reason())
			}
		}

		// New buffer should receive broadcasts
		event := New(JoinEvent, int64(42), "new-buffer-test")
		hub.Broadcast(event)

		received, err := newBuffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error on new buffer: '%v'", err)
		}
		if received.Trace() != event.Trace() {
			t.Errorf("expected trace '%s', got: '%s'", event.Trace(), received.Trace())
		}
	})

	t.Run("after close", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		hub.Close(FinishReason)

		buffer := NewBuffer(8)
		defer buffer.Close(0)

		err := hub.Register("player-1", buffer)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrHubClosed {
			t.Errorf("expected ErrHubClosed, got: '%v'", err)
		}
	})
}

func TestDeregister(t *testing.T) {
	t.Parallel()

	t.Run("existing buffer", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		buffer := NewBuffer(8)
		defer buffer.Close(0)

		if err := hub.Register("player-1", buffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		hub.Deregister("player-1")

		// Broadcast should not reach deregistered buffer
		hub.Broadcast(New(JoinEvent, int64(42), "should-not-receive"))

		// Buffer should still be open (not closed by Deregister)
		if err := buffer.Send(New(PingEvent, int64(42), "")); err != nil {
			t.Errorf("deregistered buffer should remain open, got error: '%v'", err)
		}
	})

	t.Run("non-existent ID", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		// Should not panic
		hub.Deregister("non-existent")
	})

	t.Run("buffer not closed", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		buffer := NewBuffer(8)

		if err := hub.Register("player-1", buffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		hub.Deregister("player-1")

		// Verify buffer is still functional
		event := New(StateEvent, int64(42), "test")
		if err := buffer.Send(event); err != nil {
			t.Fatalf("buffer should remain open after deregister: '%v'", err)
		}

		received, err := buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if received.Trace() != event.Trace() {
			t.Errorf("expected trace '%s', got: '%s'", event.Trace(), received.Trace())
		}

		buffer.Close(0)
	})
}

func TestBroadcast(t *testing.T) {
	t.Parallel()

	t.Run("empty hub", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		// Should not panic
		hub.Broadcast(New(JoinEvent, int64(42), "test"))
	})

	t.Run("single buffer", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		buffer := NewBuffer(8)
		if err := hub.Register("player-1", buffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		event := New(InsertEvent, int64(42), "single-test")
		hub.Broadcast(event)

		received, err := buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if received.Trace() != event.Trace() {
			t.Errorf("expected trace '%s', got: '%s'", event.Trace(), received.Trace())
		}
	})

	t.Run("multiple buffers", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		const numBuffers = 10
		buffers := make([]*buffer, numBuffers)

		for i := range buffers {
			buffers[i] = NewBuffer(8)
			if err := hub.Register(string(rune('a'+i)), buffers[i]); err != nil {
				t.Fatalf("unexpected register error: '%v'", err)
			}
		}

		event := New(StateEvent, int64(42), "multi-test")
		hub.Broadcast(event)

		for i, buf := range buffers {
			received, err := buf.Receive()
			if err != nil {
				t.Fatalf("buffer %d: unexpected receive error: '%v'", i, err)
			}
			if received.Trace() != event.Trace() {
				t.Errorf("buffer %d: expected trace '%s', got: '%s'", i, event.Trace(), received.Trace())
			}
		}
	})

	t.Run("closed buffer cleanup", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		openBuffer := NewBuffer(8)
		closedBuffer := NewBuffer(8)

		if err := hub.Register("open", openBuffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}
		if err := hub.Register("closed", closedBuffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		// Close one buffer before broadcast
		closedBuffer.Close(IdleReason)

		// Broadcast should succeed and clean up closed buffer
		event := New(JoinEvent, int64(42), "cleanup-test")
		hub.Broadcast(event)

		// Open buffer should receive the event
		received, err := openBuffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if received.Trace() != event.Trace() {
			t.Errorf("expected trace '%s', got: '%s'", event.Trace(), received.Trace())
		}

		// Re-register with same ID should work (proves cleanup happened)
		newBuffer := NewBuffer(8)
		defer newBuffer.Close(0)
		if err := hub.Register("closed", newBuffer); err != nil {
			t.Fatalf("unexpected register error after cleanup: '%v'", err)
		}
	})
}

func TestSend(t *testing.T) {
	t.Parallel()

	t.Run("non-existent ID", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		// Should not panic
		hub.Send("non-existent", New(JoinEvent, int64(42), "test"))
	})

	t.Run("registered buffer", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		buffer := NewBuffer(8)
		if err := hub.Register("player-1", buffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		event := New(InsertEvent, int64(42), "send-test")
		hub.Send("player-1", event)

		received, err := buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if received.Trace() != event.Trace() {
			t.Errorf("expected trace '%s', got: '%s'", event.Trace(), received.Trace())
		}
	})

	t.Run("multiple buffers targeted", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		buffer1 := NewBuffer(8)
		buffer2 := NewBuffer(8)

		if err := hub.Register("player-1", buffer1); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}
		if err := hub.Register("player-2", buffer2); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		event := New(PingEvent, int64(42), "targeted-send")
		hub.Send("player-1", event)

		// player-1 should receive the event
		received, err := buffer1.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error: '%v'", err)
		}
		if received.Trace() != event.Trace() {
			t.Errorf("expected trace '%s', got: '%s'", event.Trace(), received.Trace())
		}

		// player-2 should NOT receive the event
		select {
		case <-buffer2.events:
			t.Error("buffer2 should not have received any event")
		default:
			// Expected: no event in buffer2
		}
	})

	t.Run("closed buffer cleanup", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		defer hub.Close(0)

		buffer := NewBuffer(8)
		if err := hub.Register("player-1", buffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		// Close the buffer before sending
		buffer.Close(IdleReason)

		// Send should clean up the closed buffer
		hub.Send("player-1", New(StateEvent, int64(42), "cleanup-test"))

		// Re-register with same ID should work (proves cleanup happened)
		newBuffer := NewBuffer(8)
		defer newBuffer.Close(0)
		if err := hub.Register("player-1", newBuffer); err != nil {
			t.Fatalf("unexpected register error after cleanup: '%v'", err)
		}
	})
}

func TestClose(t *testing.T) {
	t.Parallel()

	t.Run("empty hub", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		// Should not panic
		hub.Close(FinishReason)
	})

	t.Run("with buffers", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()

		const numBuffers = 5
		buffers := make([]*buffer, numBuffers)

		for i := range buffers {
			buffers[i] = NewBuffer(8)
			if err := hub.Register(string(rune('a'+i)), buffers[i]); err != nil {
				t.Fatalf("unexpected register error: '%v'", err)
			}
		}

		reason := FinishReason
		hub.Close(reason)

		// All buffers should be closed with the provided reason
		for i, buf := range buffers {
			err := buf.Send(New(PingEvent, int64(42), ""))
			if err == nil {
				t.Errorf("buffer %d: expected error, got nil", i)
				continue
			}
			if closeErr, ok := err.(*BufferClosedError); !ok {
				t.Errorf("buffer %d: expected BufferClosedError, got: '%v'", i, err)
			} else {
				if closeErr.Reason() != reason {
					t.Errorf("buffer %d: expected reason %d, got: %d", i, reason, closeErr.Reason())
				}
			}
		}
	})

	t.Run("subsequent register", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()
		hub.Close(FinishReason)

		buffer := NewBuffer(8)
		defer buffer.Close(0)

		err := hub.Register("player-1", buffer)
		if err != ErrHubClosed {
			t.Errorf("expected ErrHubClosed, got: '%v'", err)
		}
	})

	t.Run("idempotent close", func(t *testing.T) {
		t.Parallel()

		hub := NewHub()

		buffer := NewBuffer(8)
		if err := hub.Register("player-1", buffer); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}

		// Double close should not panic
		hub.Close(FinishReason)
		hub.Close(TimeoutReason)

		// Buffer should still report the first close reason
		err := buffer.Send(New(PingEvent, int64(42), ""))
		if closeErr, ok := err.(*BufferClosedError); ok {
			if closeErr.Reason() != FinishReason {
				t.Errorf("expected first close reason %d, got: %d", FinishReason, closeErr.Reason())
			}
		}
	})
}

func TestConcurrentRegisterDeregister(t *testing.T) {
	t.Parallel()

	const (
		workers    = 8
		iterations = 100
	)

	hub := NewHub()
	defer hub.Close(0)

	var wg sync.WaitGroup

	for id := range workers {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range iterations {
				buffer := NewBuffer(8)
				key := string(rune('a' + id))

				if err := hub.Register(key, buffer); err != nil {
					t.Errorf("worker %d, iteration %d: unexpected register error: '%v'", id, i, err)
					return
				}

				hub.Deregister(key)
				buffer.Close(0)
			}
		}(id)
	}

	wg.Wait()
}

func TestConcurrentBroadcast(t *testing.T) {
	t.Parallel()

	const (
		numBuffers   = 8
		broadcasters = 4
		broadcasts   = 50
	)

	hub := NewHub()
	defer hub.Close(0)

	buffers := make([]*buffer, numBuffers)
	for i := range buffers {
		buffers[i] = NewBuffer(256)
		if err := hub.Register(string(rune('a'+i)), buffers[i]); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}
	}

	var wg sync.WaitGroup

	// Start concurrent broadcasters
	for id := range broadcasters {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range broadcasts {
				event := New(StateEvent, int64(id*broadcasts+i), "concurrent")
				hub.Broadcast(event)
			}
		}(id)
	}

	wg.Wait()

	// Each buffer should have received all broadcasts
	expectedCount := broadcasters * broadcasts
	for i, buf := range buffers {
		count := 0
		for {
			select {
			case <-buf.events:
				count++
			default:
				goto done
			}
		}
	done:
		if count != expectedCount {
			t.Errorf("buffer %d: expected %d events, got: %d", i, expectedCount, count)
		}
	}
}

func TestConcurrentSend(t *testing.T) {
	t.Parallel()

	const (
		numBuffers = 8
		senders    = 4
		sends      = 50
	)

	hub := NewHub()
	defer hub.Close(0)

	buffers := make([]*buffer, numBuffers)
	for i := range buffers {
		buffers[i] = NewBuffer(256)
		if err := hub.Register(string(rune('a'+i)), buffers[i]); err != nil {
			t.Fatalf("unexpected register error: '%v'", err)
		}
	}

	var wg sync.WaitGroup

	// Start concurrent senders targeting specific buffers
	for id := range senders {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			targetId := string(rune('a' + (id % numBuffers)))
			for i := range sends {
				event := New(StateEvent, int64(id*sends+i), "concurrent")
				hub.Send(targetId, event)
			}
		}(id)
	}

	wg.Wait()

	// Verify each buffer received events only from its targeted senders
	for i, buf := range buffers {
		count := 0
		for {
			select {
			case <-buf.events:
				count++
			default:
				goto done
			}
		}
	done:
		// Calculate expected events for this buffer
		expectedSenders := 0
		for id := range senders {
			if id%numBuffers == i {
				expectedSenders++
			}
		}
		expectedCount := expectedSenders * sends

		if count != expectedCount {
			t.Errorf("buffer %d: expected %d events, got: %d", i, expectedCount, count)
		}
	}
}

func TestUnderloadHub(t *testing.T) {
	t.Run("register during close", func(t *testing.T) {
		for i := range 1000 {
			hub := NewHub()

			var wg sync.WaitGroup
			ready := make(chan struct{})

			const numRegisterers = 10

			for range numRegisterers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-ready
					buffer := NewBuffer(8)
					err := hub.Register("contested-id", buffer)
					// Valid outcomes: nil or ErrHubClosed
					if err != nil && err != ErrHubClosed {
						t.Errorf("iteration %d: unexpected error: '%v'", i, err)
					}
					buffer.Close(0)
				}()
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ready
				hub.Close(FinishReason)
			}()

			close(ready)
			wg.Wait()
		}
	})

	t.Run("broadcast during close", func(t *testing.T) {
		for i := range 1000 {
			hub := NewHub()

			// Pre-register some buffers
			buffers := make([]*buffer, 5)
			for j := range buffers {
				buffers[j] = NewBuffer(256)
				hub.Register(string(rune('a'+j)), buffers[j])
			}

			var wg sync.WaitGroup
			ready := make(chan struct{})

			const numBroadcasters = 10

			for range numBroadcasters {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-ready
					for range 100 {
						hub.Broadcast(New(StateEvent, int64(42), ""))
					}
				}()
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ready
				hub.Close(FinishReason)
			}()

			close(ready)
			wg.Wait()

			// Verify no panic occurred - test passes if we reach here
			_ = i
		}
	})

	t.Run("register deregister broadcast", func(t *testing.T) {
		for i := range 500 {
			hub := NewHub()

			var wg sync.WaitGroup
			ready := make(chan struct{})

			const workers = 8

			// Registerers
			for id := range workers {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					<-ready
					for range 50 {
						buffer := NewBuffer(8)
						hub.Register(string(rune('a'+id)), buffer)
						buffer.Close(0)
					}
				}(id)
			}

			// Deregisterers
			for id := range workers {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					<-ready
					for range 50 {
						hub.Deregister(string(rune('a' + id)))
					}
				}(id)
			}

			// Broadcasters
			for range workers {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-ready
					for range 50 {
						hub.Broadcast(New(StateEvent, int64(42), ""))
					}
				}()
			}

			close(ready)
			wg.Wait()
			hub.Close(0)

			// Verify no panic occurred
			_ = i
		}
	})

	t.Run("send during close", func(t *testing.T) {
		for i := range 1000 {
			hub := NewHub()

			// Pre-register some buffers
			buffers := make([]*buffer, 5)
			for j := range buffers {
				buffers[j] = NewBuffer(256)
				hub.Register(string(rune('a'+j)), buffers[j])
			}

			var wg sync.WaitGroup
			ready := make(chan struct{})

			const numSenders = 10

			for id := range numSenders {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					<-ready
					targetId := string(rune('a' + (id % 5)))
					for range 100 {
						hub.Send(targetId, New(StateEvent, int64(42), ""))
					}
				}(id)
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ready
				hub.Close(FinishReason)
			}()

			close(ready)
			wg.Wait()

			// Verify no panic occurred
			_ = i
		}
	})

	t.Run("register deregister send", func(t *testing.T) {
		for i := range 500 {
			hub := NewHub()

			var wg sync.WaitGroup
			ready := make(chan struct{})

			const workers = 8

			// Registerers
			for id := range workers {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					<-ready
					for range 50 {
						buffer := NewBuffer(8)
						hub.Register(string(rune('a'+id)), buffer)
						buffer.Close(0)
					}
				}(id)
			}

			// Deregisterers
			for id := range workers {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					<-ready
					for range 50 {
						hub.Deregister(string(rune('a' + id)))
					}
				}(id)
			}

			// Senders
			for id := range workers {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					<-ready
					targetId := string(rune('a' + (id % workers)))
					for range 50 {
						hub.Send(targetId, New(StateEvent, int64(42), ""))
					}
				}(id)
			}

			close(ready)
			wg.Wait()
			hub.Close(0)

			// Verify no panic occurred
			_ = i
		}
	})
}
