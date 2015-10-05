package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/partkyle/cptplanet"
	//"github.com/sendgrid/ln"
)

// app name
const APP = "poster"

// event payload
type EventPayload struct {
	UserID int             `json:"user_id"`
	Event  json.RawMessage `json:"event"`
}

func main() {
	var wg sync.WaitGroup

	// configs
	envSettings := cptplanet.Settings{Prefix: "POSTER_", ErrorOnExtraKeys: false, ErrorOnMissingKeys: true, ErrorOnParseErrors: true}
	env := cptplanet.NewEnvironment(envSettings)
	redisServer := env.String("REDIS_SERVER", "127.0.0.1:6379", "host to bind")
	err := env.Parse()
	if err != nil {
		log.Fatalf("could not get required configs: %v", err)
	}

	var logger *syslog.Writer
	logger, err = syslog.New(syslog.LOG_DEBUG|syslog.LOG_LOCAL0, APP)
	if err != nil {
		log.Fatalf("could not involke syslog logger: %v", err)
	}

	multiplexer := NewMultiplexer()
	multiplexer.SetRedisPool(createRedisPool(*redisServer))
	multiplexer.SetPipeWorkerFactory(WorkerFactory)
	multiplexer.SetRejectPipe(createRejectPipe())
	multiplexer.Start()

	scanner := bufio.NewScanner(os.Stdin)

	logger.Info(fmt.Sprintf("%s started\n", APP))

	//TODO: EOF should terminate the program
	for scanner.Scan() {
		m := &PipeMessage{
			payload: scanner.Bytes(),
		}
		m.SetAckHandler(func() error {
			logger.Debug("Ack called\n")
			wg.Done()
			return nil
		})
		wg.Add(1)

		multiplexer.In() <- m
		logger.Debug("sent multiplex msg")
	}
	if err := scanner.Err(); err != nil {
		logger.Err(fmt.Sprintf("error reading standard input: %s", err.Error()))
	}

	wg.Wait()
	// need to flush rejected logs because it's buffered
	multiplexer.GetRejectPipe().Stop()
	logger.Err(fmt.Sprintf("%s shutting down\n", APP))
}

func createRejectPipe() Pipe {
	rejectLogger, err := syslog.New(syslog.LOG_DEBUG|syslog.LOG_LOCAL1, APP)
	if err != nil {
		log.Fatalf("could not involke syslog logger: %v", err)
	}
	//logger := ln.New("syslog", "DEBUG", "LOCAL1", APP)
	rejecter := NewRejecter()
	rejecter.SetLogger(rejectLogger)
	rejecter.Start()
	return rejecter
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

func createRedisPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
