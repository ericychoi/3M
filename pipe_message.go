package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type PipeMessage struct {
	id            string
	payload       []byte
	ackHandler    func() error
	rejectHandler func(error)
	metadata      map[string]string
	redisPool     *redis.Pool
}

func (p *PipeMessage) Ack() error {
	fmt.Println("calling ack()")
	return p.ackHandler()
}

func (p *PipeMessage) Reject(e error) {
	p.rejectHandler(e)
	return
}

func (p *PipeMessage) ID() string {
	return p.id
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
func (p *PipeMessage) SaveMetadata() {
	fmt.Println("calling SaveMetadata()")

	conn := p.redisPool.Get()
	defer conn.Close()
	var byteMetadata []byte
	if _, err := conn.Do("SET", "poster.message.metadata."+p.id, byteMetadata); err != nil {
		fmt.Println("error in redis saving in SaveMetadata()")
	}
}
