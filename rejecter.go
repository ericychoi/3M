package main

import (
	"bytes"
	//	"github.com/sendgrid/ln"
	"log"
	"log/syslog"
)

type Rejecter struct {
	in     chan Message
	out    chan Message
	logger *syslog.Writer
}

// for now, just proxy
func (r *Rejecter) In() chan<- Message {
	return r.in
}

func (r *Rejecter) Out() <-chan Message {
	return r.out
}

func (r *Rejecter) SetLogger(l *syslog.Writer) {
	r.logger = l
}

func (r *Rejecter) GetLogger() *syslog.Writer {
	return r.logger
}

//TODO
func (r *Rejecter) Start() {
	go func() {
		for {
			m := <-r.in
			payload := m.Payload()
			log.Printf("Rejecter: payload: %s\n", payload)
			buf := &bytes.Buffer{}
			buf.Write(payload)
			r.logger.Err(buf.String())
			m.Ack()
		}
	}()
}

func (r *Rejecter) Stop() {
	r.logger.Close()
}

func NewRejecter() *Rejecter {
	return &Rejecter{
		in:  make(chan Message),
		out: make(chan Message),
	}
}
