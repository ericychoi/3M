package main

type Poster struct {
	in  <-chan []byte
	out chan<- []byte
}

func (m *Poster) In() <-chan []byte {
	return in
}

// it doesn't output anything
func (m *Poster) Out() chan<- []byte {
	return out
}

func (m *Poster) Start() {
	go func() {
		for {
			select {
			case payload := <-m.in:
				//TODO: print out
			}
		}
	}()
}

func (m *Poster) Stop() {}

func NewPoster() *Poster {
	return &Poster{
		in:  make(<-chan []byte),
		out: make(chan<- []byte),
	}
}
