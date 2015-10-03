package main

type Message interface {
	Payload() []byte
	Metadata() map[string]interface{}
	Ack() error
	Reject(error)
	ID() string
}
