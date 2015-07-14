package main

type Pipe interface {
	In() <-chan []byte
	Out() chan<- []byte
	Start()
	Stop()
}
