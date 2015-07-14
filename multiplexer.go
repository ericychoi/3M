package main

import (
	"encoding/json"
	"log"
)

type pipeWorkerFactory func(chan<- []byte)

type Multiplexer struct {
	in  <-chan []byte
	out chan<- []byte
	pipeFactory pipeWorkerFactory
	var workerMap map[int]Pipe
}

func (m *Multiplexer) In() <-chan []byte {
	return in
}

// it doesn't output anything
func (m *Multiplexer) Out() chan<- []byte {
	return nil
}

func (m *Multiplexer) Start() {
	go func() {
		var eventPayloadBatch []EventPayload
		for {
			select {
			case batchRaw := <-m.in:
				if err := json.Unmarshal(batchRaw, eventPayloadBatch); err != nil {
					log.Printf("error parsing event batch: %s", err.Error())
				}
				for _, payload := range eventPayloadBatch {
					if payload.UserID == 0 {
						log.Printf("error parsing event payload, couldn't get user_id: %+v", payload)
						continue
					}

					var worker Pipe
					if m.workerMap[payload.UserID] == nil {
						worker = m.pipeFactory()
						workerMap[payload.UserID] = worker
						worker.Start()
					} else {
						worker = workerMap[payload.UserID]
					}

					worker.In() <- scanner
				}

			}

		}
	}()
}

func (m *Multiplexer) Stop() {}

func NewMultiplexer() *Multiplexer {
	return &Multiplexer{
		in:  make(<-chan []byte),
		out: nil,
	}
}
