package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Poster struct {
	in  chan Message
	out chan Message
}

func (p *Poster) In() chan<- Message {
	return p.in
}

func (p *Poster) SetIn(c chan Message) {
	p.in = c
}

// it doesn't output anything
func (p *Poster) Out() <-chan Message {
	return nil
}

func (p *Poster) Start() {
	go func() {
		var eventPayload EventPayload
		for {
			m := <-p.in
			payload := m.Payload()
			log.Printf("payload: %s\n", payload)
			if err := json.Unmarshal(payload, &eventPayload); err != nil {
				log.Printf("poster: error parsing event: %s", err.Error())
				continue
			}
			if eventPayload.UserID == 0 {
				log.Printf("poster: error parsing event payload, couldn't get user_id: %+v", payload)
				continue
			}
			//TODO post based on userID ...
			url := "http://localhost:8000"
			req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
			req.Header.Set("X-Custom-Header", "myvalue")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)

			if err != nil {
				log.Printf("poster: error posting event: %s", err.Error())
				throttleMessage(m)
				m.Reject(err)
				continue
			} else if resp.StatusCode/100 != 2 {
				log.Printf("poster: status code non-2xx: %d", resp.StatusCode)
				throttleMessage(m)
				m.Reject(err)
				continue
			}

			log.Printf("response %s", resp.Status)
			resp.Body.Close()
			m.Ack()
		}
	}()
}

func (p *Poster) Stop() {}

// write to messsage's metadata that this needs to be throttled
func throttleMessage(m Message) {
	metadata := m.Metadata()
	currentDelay, err := strconv.Atoi(metadata[THROTTLER_DELAY])
	if err != nil {
		log.Printf(
			"poster: could not get throttler_next_attempt from metadata. got %s", metadata[THROTTLER_DELAY],
		)
		currentDelay = 0
	}
	nextDelay := currentDelay * 2
	log.Printf("poster: setting next delay to %d", nextDelay)
	metadata[THROTTLER_DELAY] = strconv.Itoa(nextDelay)
}

func NewPoster() *Poster {
	return &Poster{
		in:  make(chan Message),
		out: nil,
	}
}
