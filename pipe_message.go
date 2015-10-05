package main

import "github.com/garyburd/redigo/redis"

type PipeMessage struct {
	payload       []byte
	ackHandler    func() error
	rejectHandler func(error)
	metadata      map[string]string
}

func (p *PipeMessage) Ack() error {
	p.SaveMetadata()
	return p.ackHandler()
}

func (p *PipeMessage) Reject(e error) {
	p.SaveMetadata()
	p.rejectHandler(e)
	return
}

func (p *PipeMessage) ID() string {
	return "id"
}

func (p *PipeMessage) Payload() []byte {
	return p.payload
}

//TODO: store metadata
func (p *PipeMessage) Metadata() map[string]string {
	if p.metadata == nil {
		p.metadata = make(map[string]string)
	}
	return p.metadata
}

func (p *PipeMessage) SetAckHandler(f func() error) {
	p.ackHandler = f
}

func (p *PipeMessage) SetRejectHandler(f func(error)) {
	p.rejectHandler = f
}

//TODO: make it an interface so that it can use any backend storage
//for now, this is married to redis
func (p *PipeMessage) SaveMetadata(f func(error)) {

}
