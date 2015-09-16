package main

type Pipe interface {
	In() chan<- Message
	Out() <-chan Message
	Start()
	Stop()
}
