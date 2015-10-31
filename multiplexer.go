package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/garyburd/redigo/redis"
)

type pipeWorkerFactory func() Pipe

// a terminal pipe (messages ends here (ack/reject)); creates PipeMessages, uses pipeFactory to
// create multiple pipes and send the PipeMessages
// when pipeMessage gets rejected it goes to rejectPipe
type Multiplexer struct {
	in          chan Message
	out         chan Message
	pipeFactory pipeWorkerFactory
	rejectPipe  Pipe
	workerMap   map[int]Pipe
	redisPool   *redis.Pool
}

func (m *Multiplexer) In() chan<- Message {
	return m.in
}

// it doesn't output anything
func (m *Multiplexer) Out() <-chan Message {
	return nil
}

// event looks like
/*
{
    "user_id": 180,
    "event": {
      "sg_event_id": "LhVSEIH-RyqLovf1-i1xLA",
      "sg_message_id": "core-development-build.22201.54FDEBF41.0",
      "abcd": "800",
      "event": "processed",
      "email": "eric.choi099@sendgrid.com",
      "smtp-id": "<1425927231.9775453593371313@core-development-build>",
      "timestamp": 1425927232,
      "_sg_data_": {
        "appid": 1
      }
    }
}
*/
func (m *Multiplexer) Start() {
	go func() {
		var payload EventPayload
		for {
			msg := <-m.in
			log.Printf("received: %#v\n", msg)

			batchRaw := msg.Payload()

			if err := json.Unmarshal(batchRaw, &payload); err != nil {
				log.Printf("error parsing event batch: %s, %s", err.Error(), batchRaw)
				continue
			}
			log.Printf("eventPayloadBatch: %#v\n", payload)

			var wg sync.WaitGroup
			go func() {
				wg.Wait()
				msg.Ack()
			}()

			if payload.UserID == 0 {
				log.Printf("error parsing event payload, couldn't get user_id: %+v", payload)
				continue
			}
			var worker Pipe
			if m.workerMap[payload.UserID] == nil {
				log.Printf("creating")
				m.workerMap[payload.UserID] = m.pipeFactory()
				m.workerMap[payload.UserID].Start()
			}

			worker = m.workerMap[payload.UserID]
			log.Printf("about to send event payload %#v\n", payload)

			jsonPayload, _ := json.Marshal(&payload)

			// create a new message (maybe factor this out to a factory?)
			splitMsg := &PipeMessage{
				payload:   jsonPayload,
				redisPool: m.redisPool,
			}
			splitMsg.SetAckHandler(func() error {
				log.Printf("Ack called on splitMsg\n")
				splitMsg.SaveMetadata()
				wg.Done()
				return nil
			})
			splitMsg.SetRejectHandler(func(error) {
				log.Printf("Reject called on splitMsg\n")
				splitMsg.SaveMetadata()
				m.rejectPipe.In() <- splitMsg
				return
			})

			wg.Add(1)
			worker.In() <- splitMsg
			log.Printf("sent worker msg")
		}
	}()
}

// TODO
func (m *Multiplexer) Stop() {}

func (m *Multiplexer) SetPipeWorkerFactory(p pipeWorkerFactory) {
	m.pipeFactory = p
}

func (m *Multiplexer) SetRejectPipe(p Pipe) {
	m.rejectPipe = p
}

func (m *Multiplexer) GetRejectPipe() Pipe {
	return m.rejectPipe
}

func (m *Multiplexer) SetRedisPool(p *redis.Pool) {
	m.redisPool = p
}

func NewMultiplexer() *Multiplexer {
	return &Multiplexer{
		in:        make(chan Message),
		out:       nil,
		workerMap: make(map[int]Pipe),
	}
}
