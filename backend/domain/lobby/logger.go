package lobby

import (
	"sync"
)

type logMsg struct {
	Row    int
	Column int
	Value  int
	Player string
	Time   int64
}

type Logger struct {
	msgs    chan *logMsg
	storage []*logMsg
	lock    sync.RWMutex
	done    chan struct{}
}

func newLogger() *Logger {
	return &Logger{
		msgs: make(chan *logMsg, 1024),
		done: make(chan struct{}),
	}
}

func (l *Lobby) log(msg *logMsg) {
	l.lock.RLock()
	select {
	case <-l.done:
		return // lobby is closed
	default:
	}
	l.logger.msgs <- msg
	l.lock.RUnlock()
}

func (l *Logger) storer() {
	for msg := range l.msgs {
		l.lock.Lock()
		l.storage = append(l.storage, msg)
		l.lock.Unlock()
	}
	close(l.done)
}
