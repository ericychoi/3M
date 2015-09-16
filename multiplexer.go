package main

import (
	"encoding/json"
	"log"
	"sync"
)

type pipeWorkerFactory func() Pipe

type Multiplexer struct {
	in          chan Message
	out         chan Message
	pipeFactory pipeWorkerFactory
	workerMap   map[int]Pipe
}

func (m *Multiplexer) In() chan<- Message {
	return m.in
}

// it doesn't output anything
func (m *Multiplexer) Out() <-chan Message {
	return nil
}

// event batch looks like
/*[
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
  },
  {
    "user_id": 180,
    "event": {
      "sg_event_id": "m-ToYAJ8TpazjKIEdunpKA",
      "sg_message_id": "core-development-build.22201.54FDEC3F1.0",
      "abcd": "800",
      "event": "processed",
      "email": "eric.choi099@sendgrid.com",
      "smtp-id": "<1425927232.1513133910567075@core-development-build>",
      "timestamp": 1425927232,
      "_sg_data_": {
        "appid": 1
      }
    }
  }
]
*/
func (m *Multiplexer) Start() {
	go func() {
		var eventPayloadBatch []EventPayload
		for {
			//batchRaw := <-m.in
			msg := <-m.in
			log.Printf("received: %#v\n", msg)

			batchRaw := msg.Payload()

			if err := json.Unmarshal(batchRaw, &eventPayloadBatch); err != nil {
				log.Printf("error parsing event batch: %s", err.Error())
				continue
			}
			log.Printf("eventPayloadBatch: %#v\n", eventPayloadBatch)

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				wg.Wait()
				msg.Ack()
			}()

			for _, payload := range eventPayloadBatch {
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
				splitMsg := &PipeMessage{payload: jsonPayload}
				splitMsg.SetAckHandler(func() error {
					log.Printf("Ack called on splitMsg\n")
					wg.Done()
					return nil
				})
				wg.Add(1)
				worker.In() <- splitMsg
				log.Printf("sent worker msg")
			}
			wg.Done()
		}
	}()
}

// TODO
func (m *Multiplexer) Stop() {}

func (m *Multiplexer) SetPipeWorkerFactory(p pipeWorkerFactory) {
	m.pipeFactory = p
}

func NewMultiplexer() *Multiplexer {
	return &Multiplexer{
		in:        make(chan Message),
		out:       nil,
		workerMap: make(map[int]Pipe),
	}
}
