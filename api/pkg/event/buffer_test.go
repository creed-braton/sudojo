package event

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestBuffer(t *testing.T) {
	t.Run("send and receive event", func(t *testing.T) {
		buffer := NewBuffer(256)
		defer buffer.Close(0)

		want := "test-message"
		var e Event = New("", int64(42), want)
		if err := buffer.Send(e); err != nil {
			t.Fatalf("unexpected send error: '%v'", err)
		}

		e, err := buffer.Receive()
		if err != nil {
			t.Fatalf("unexpected receive error event: '%v'", err)
		}

		if e.Trace() != want {
			t.Errorf("expected: '%s', got: '%s'", want, e.Trace())
		}
	})

	t.Run("closed event buffer", func(t *testing.T) {
		reason := 1
		buffer := NewBuffer(256)
		buffer.Close(reason)

		err := buffer.Send(New("", int64(42), ""))
		if err, ok := err.(*BufferClosedError); !ok {
			t.Errorf("expected BufferClosedError on send, got: '%v'", err)
		} else {
			if err.Reason() != reason {
				t.Errorf("expected close reason %d, got: %d", reason, err.Reason())
			}
		}

		_, err = buffer.Receive()
		if err, ok := err.(*BufferClosedError); !ok {
			t.Errorf("expected BufferClosedError on receive, got: '%v'", err)
		} else {
			if err.Reason() != reason {
				t.Errorf("expected close reason %d, got: %d", reason, err.Reason())
			}
		}
	})

	t.Run("full event buffer", func(t *testing.T) {
		size := 256
		buffer := NewBuffer(size)

		for range size {
			e := New("", int64(42), "")
			if err := buffer.Send(e); err != nil {
				t.Fatalf("unexpected send error: '%v'", err)
			}
		}

		err := buffer.Send(New("", int64(42), ""))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err != ErrBufferFull {
			t.Errorf("expected receive error '%v', got: '%v'", ErrBufferFull, err)
		}
	})

	t.Run("concurrent senders", func(t *testing.T) {
		const (
			sender = 8
			num    = 32
			total  = sender * num
		)

		buffer := NewBuffer(256)
		defer buffer.Close(0)

		var wg sync.WaitGroup

		// Start concurrent senders
		for id := range sender {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := range num {
					msg := fmt.Sprintf("%d-%d", id, i)
					e := New("", int64(42), msg)
					if err := buffer.Send(e); err != nil {
						t.Errorf("unexpected send error from sender %d: '%v'", id, err)
						return
					}
				}
			}(id)
		}

		received := make(map[string]bool)
		for range total {
			e, err := buffer.Receive()
			if err != nil {
				t.Fatalf("unexpected receive error: '%v'", err)
			}

			if received[e.Trace()] {
				t.Errorf("duplicate message received: '%s'", e.Trace())
			}
			received[e.Trace()] = true
		}

		wg.Wait()

		for id := range sender {
			for i := range num {
				msg := fmt.Sprintf("%d-%d", id, i)
				if !received[msg] {
					t.Errorf("message missing: '%s'", msg)
				}
			}
		}
	})
}

func TestUnderload(t *testing.T) {
	t.Run("send during close", func(t *testing.T) {
		for i := range 1000 {
			buffer := NewBuffer(256)

			var wg sync.WaitGroup
			ready := make(chan struct{})

			const numSenders = 10

			// Start senders that wait for the signal
			for range numSenders {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-ready // Wait for signal
					for range 100 {
						err := buffer.Send(New("", int64(42), ""))
						// Only valid outcomes: success, buffer full, or buffer closed
						if err != nil && err != ErrBufferFull && !errors.As(err, new(*BufferClosedError)) {
							t.Errorf("iteration %d: unexpected error: '%v'", i, err)
						}
					}
				}()
			}

			// Start closer that waits for the signal
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ready // Wait for signal
				buffer.Close(0)
			}()

			// Release all goroutines simultaneously to maximize race potential
			close(ready)
			wg.Wait()
		}
	})

	t.Run("concurrent receives with close reason", func(t *testing.T) {
		for iteration := range 100 {
			const (
				numReceivers = 10
				reason       = 42
			)

			buffer := NewBuffer(256)
			var wg sync.WaitGroup
			ready := make(chan struct{})

			// Pre-populate buffer with some events so receivers have work to do
			for i := range 5 {
				buffer.Send(New("", int64(i), ""))
			}

			// Start concurrent receivers
			for id := range numReceivers {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					<-ready
					for {
						_, err := buffer.Receive()
						if err == nil {
							continue
						}
						if err, ok := err.(*BufferClosedError); !ok {
							t.Errorf("iteration %d, receiver %d: expected BufferClosedError, got: '%v'", iteration, id, err)
						} else {
							if err.Reason() != reason {
								t.Errorf("iteration %d, receiver %d: expected close reason %d, got: %d", iteration, id, reason, err.Reason())
							}
						}
						return
					}
				}(id)
			}

			// Start closer
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-ready
				buffer.Close(reason)
			}()

			// Release all goroutines simultaneously to maximize race potential
			close(ready)
			wg.Wait()
		}
	})
}
