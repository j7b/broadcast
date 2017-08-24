// Package broadcast implements
// concurrent broadcasts to multiple goroutines.
package broadcast

import "sync"

// BroadcastInterface implementations are sent by Senders and
// received by Receivers. Data races are only possible if
// implementations are (or contain) reference types. Broadcast()
// is never executed by this package; it is an identification-only method.
type BroadcastInterface interface {
	Broadcast()
}

// Sender sends Broadcasts to Receivers and generates Receivers.
type Sender interface {
	Send(BroadcastInterface)
	Receiver() Receiver
}

// Receiver receives Broadcasts from a Sender. Receive will
// block until Chan() is closed by the Sender.
type Receiver interface {
	Chan() <-chan struct{}
	Receive() BroadcastInterface
}

type broadcast struct {
	c    chan struct{}
	val  BroadcastInterface
	next *broadcast
}

type container struct {
	b *broadcast
}

func (c *container) Send(b BroadcastInterface) {
	c.b = c.b.send(b)
}

func (c *container) Chan() <-chan struct{} {
	return c.b.c
}

func (c *container) Receiver() Receiver {
	return &container{c.b}
}

func (c *container) Receive() (b BroadcastInterface) {
	b, c.b = c.b.receive()
	return
}

func (b *broadcast) receive() (BroadcastInterface, *broadcast) {
	<-b.c
	return b.val, b.next
}

func (b *broadcast) send(bc BroadcastInterface) *broadcast {
	b.val = bc
	n := &broadcast{c: make(chan struct{})}
	b.next = n
	close(b.c)
	return n
}

// New creates a new Sender.
func New() Sender {
	return &container{&broadcast{c: make(chan struct{})}}
}

type sharedSender struct {
	m sync.Mutex
	*container
}

// SharedSender is a Sender that manages
// its own state and is safe for concurrent
// use by multiple goroutines. "Transceivers"
// sharing a SharedSender should call
// Receiver() and prepare to service the Receiver
// before calling Send in most cases.
type SharedSender interface {
	Send(b BroadcastInterface)
	Receiver() Receiver
}

// NewShared returns a SharedSender.
func NewShared() SharedSender {
	return &sharedSender{container: &container{&broadcast{c: make(chan struct{})}}}
}

func (s *sharedSender) Send(b BroadcastInterface) {
	s.m.Lock()
	defer s.m.Unlock()
	s.container.Send(b)
}

func (s *sharedSender) Receiver() Receiver {
	s.m.Lock()
	defer s.m.Unlock()
	return s.container.Receiver()
}
