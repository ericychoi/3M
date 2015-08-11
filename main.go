package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

type EventPayload struct {
	UserID int `json:user_id`
	Event  json.RawMessage
}

func main() {
	//	outChans = make(map[int]chan []byte)
	multiplexer := NewMultiplexer()
	multiplexer.SetPipeWorkerFactory(ThreeMWorkerFactory)
	multiplexer.Start()

	log.Println("3M started!!")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		multiplexer.In() <- scanner.Bytes()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading standard input: %s", err.Error())
	}

	log.Println("3M shutting down")
}

func ThreeMWorkerFactory() Pipe {
	link := make(chan []byte)
	throttler := &Throttler{
		in:  make(chan []byte),
		out: link,
	}

	poster := &Poster{
		in:  link,
		out: nil,
	}

	p := &Pipeline{pipes: []Pipe{throttler, poster}}
	return p
}
