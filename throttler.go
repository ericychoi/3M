package main

type Throttler struct {
	in  <-chan []byte
	out chan<- []byte
}

func (m *Throttler) In() <-chan []byte {
	return in
}

// it doesn't output anything
func (m *Throttler) Out() chan<- []byte {
	return out
}

func (m *Throttler) Start() {}

func (m *Throttler) Stop() {}

func NewThrottler() *Throttler {
	return &Throttler{
		in:  make(<-chan []byte),
		out: make(chan<- []byte),
	}
}
