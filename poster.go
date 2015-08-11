package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type Poster struct {
	in  chan []byte
	out chan []byte
}

func (p *Poster) In() chan<- []byte {
	return p.in
}

func (p *Poster) SetIn(c chan []byte) {
	p.in = c
}

// it doesn't output anything
func (p *Poster) Out() <-chan []byte {
	return nil
}

func (p *Poster) Start() {
	go func() {
		var eventPayload EventPayload
		for {
			payload := <-p.in
			if err := json.Unmarshal(payload, &eventPayload); err != nil {
				log.Printf("poster: error parsing event: %s", err.Error())
				continue
			}
			if eventPayload.UserID == 0 {
				log.Printf("poster: error parsing event payload, couldn't get user_id: %+v", payload)
				continue
			}
			//TODO post based on userID ...

			url := "http://requestb.in/r2rqcvr2"
			req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
			req.Header.Set("X-Custom-Header", "myvalue")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("poster: error posting event: %s", err.Error())
				continue
			}
			defer resp.Body.Close()

			log.Printf("response %d", resp.Status)
		}
	}()
}

func (p *Poster) Stop() {}

func NewPoster() *Poster {
	return &Poster{
		in:  make(chan []byte),
		out: nil,
	}
}
