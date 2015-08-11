package main

import (
	"encoding/json"
	"log"
)

type Throttler struct {
	in  chan []byte
	out chan []byte
}

// for now, just proxy
func (t *Throttler) In() chan<- []byte {
	return t.in
}

func (t *Throttler) Out() <-chan []byte {
	return t.out
}

func (t *Throttler) Start() {
	go func() {
		var eventPayload EventPayload
		for {
			payload := <-t.in
			if err := json.Unmarshal(payload, &eventPayload); err != nil {
				log.Printf("throttler: error parsing event: %s", err.Error())
				continue
			}
			if eventPayload.UserID == 0 {
				log.Printf("throttler: error parsing event payload, couldn't get user_id: %+v", payload)
				continue
			}
			//TODO throttle based on userID ...
			t.out <- payload
		}
	}()
}

func (m *Throttler) Stop() {}

func NewThrottler() *Throttler {
	return &Throttler{
		in:  make(chan []byte),
		out: make(chan []byte),
	}
}
