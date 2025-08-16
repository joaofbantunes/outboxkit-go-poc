package polling

type OutboxTrigger interface {
	OnNewMessages()
}

// to look into later, just leaving it here to not forget
type KeyedOutboxTrigger interface{}

type OutboxListener interface {
	Chan() <-chan struct{}
}

type listener struct {
	ch chan struct{}
}

func newOutboxListener() *listener {
	return &listener{
		ch: make(chan struct{}),
	}
}

func (l *listener) Chan() <-chan struct{} {
	return l.ch
}

func (l *listener) OnNewMessages() {
	select {
	case l.ch <- struct{}{}:
	default:
		// if the channel is full, we don't block
	}
}
