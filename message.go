package main

type Message interface {
	Payload() []byte
	Metadata() map[string]string
	Ack() error
	Reject(error)
	ID() string
}
