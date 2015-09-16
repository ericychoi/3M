package main

type Message interface {
	Payload() []byte
	Ack() error
	Reject(error)
	ID() string
}
