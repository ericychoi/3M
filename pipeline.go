package main

type Pipeline struct {
	pipes []Pipe
}

func (p *Pipeline) Start() {
	for _, p := range p.pipes {
		p.Start()
	}
}

func (p *Pipeline) Stop() {
	//TODO
}

func (p *Pipeline) In() chan<- Message {
	return p.pipes[0].In()
}

// it doesn't output anything
func (p *Pipeline) Out() <-chan Message {
	return p.pipes[len(p.pipes)-1].Out()
}
