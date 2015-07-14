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
	outChans = make(map[int]chan []byte)
	multiplexer := NewMultiplexer()
	multiplexer.SetPipeWorkerFactory(&pipeWorkerFactory)
	multiplexer.Start()

	log.Println("Anemone started!!")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		multiplexer.In() <- scanner.Bytes()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("error reading standard input: %s", err.Error())
	}

	log.Println("Anemone shutting down")
}

func pipeWorkerFactory() {
	throttler := NewThrottler(make(chan []byte))
	poster := NewPoster(throttler.In())

	throttler.Start()
	poster.Start()
}
