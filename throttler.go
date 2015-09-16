package main

import (
	"encoding/json"
	"log"
)

type Throttler struct {
	in  chan Message
	out chan Message
}

// for now, just proxy
func (t *Throttler) In() chan<- Message {
	return t.in
}

func (t *Throttler) Out() <-chan Message {
	return t.out
}

func (t *Throttler) Start() {
	go func() {
		var eventPayload EventPayload
		for {
			m := <-t.in
			payload := m.Payload()
			log.Printf("throttler: payload: %s\n", payload)

			if err := json.Unmarshal(payload, &eventPayload); err != nil {
				log.Printf("throttler: error parsing event: %s", err.Error())
				continue
			}
			if eventPayload.UserID == 0 {
				log.Printf("throttler: error parsing event payload, couldn't get user_id: %+v", payload)
				continue
			}
			//TODO throttle based on userID ...
			t.out <- m
		}
	}()
}

func (m *Throttler) Stop() {}

func NewThrottler() *Throttler {
	return &Throttler{
		in:  make(chan Message),
		out: make(chan Message),
	}
}
