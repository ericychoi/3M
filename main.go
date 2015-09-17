package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"sync"
)

// app name
const APP = "Poster"

// event payload
type EventPayload struct {
	UserID int             `json:"user_id"`
	Event  json.RawMessage `json:"event"`
}

func main() {
	multiplexer := NewMultiplexer()
	multiplexer.SetPipeWorkerFactory(WorkerFactory)
	multiplexer.SetRejectPipe(createRejectPipe())
	multiplexer.Start()

	log.Printf("%s started\n", APP)

	scanner := bufio.NewScanner(os.Stdin)

	var wg sync.WaitGroup

	for scanner.Scan() {
		m := &PipeMessage{
			payload: scanner.Bytes(),
		}
		m.SetAckHandler(func() error {
			log.Printf("Ack called\n")
			wg.Done()
			return nil
		})
		wg.Add(1)

		multiplexer.In() <- m
		log.Printf("sent multiplex msg")
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading standard input: %s", err.Error())
	}

	wg.Wait()
	log.Printf("%s shutting down\n", APP)
}

func createRejectPipe() Pipe {
	//TODO
	return nil
}

func WorkerFactory() Pipe {
	link := make(chan Message)
	throttler := &Throttler{
		in:  make(chan Message),
		out: link,
	}

	poster := &Poster{
		in:  link,
		out: nil,
	}

	p := &Pipeline{pipes: []Pipe{throttler, poster}}
	return p
}
