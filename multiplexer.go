package main

import (
	"encoding/json"
	"log"
)

type pipeWorkerFactory func() Pipe

type Multiplexer struct {
	in          chan []byte
	out         chan []byte
	pipeFactory pipeWorkerFactory
	workerMap   map[int]Pipe
}

func (m *Multiplexer) In() chan<- []byte {
	return m.in
}

// it doesn't output anything
func (m *Multiplexer) Out() <-chan []byte {
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
			batchRaw := <-m.in
			if err := json.Unmarshal(batchRaw, &eventPayloadBatch); err != nil {
				log.Printf("error parsing event batch: %s", err.Error())
				continue
			}
			for _, payload := range eventPayloadBatch {
				if payload.UserID == 0 {
					log.Printf("error parsing event payload, couldn't get user_id: %+v", payload)
					continue
				}

				var worker Pipe
				if m.workerMap[payload.UserID] == nil {
					m.workerMap[payload.UserID] = m.pipeFactory()
					m.workerMap[payload.UserID].Start()
				}

				worker = m.workerMap[payload.UserID]

				worker.In() <- batchRaw
			}
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
		in:  make(chan []byte),
		out: nil,
	}
}
