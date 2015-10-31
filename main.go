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
	//"github.com/partkyle/cptplanet"
)

// app name
const APP = "poster"

// event payload
type EventPayload struct {
	UserID int             `json:"user_id"`
	Event  json.RawMessage `json:"event"`
}

var Logger *log.Logger

func main() {
	var wg sync.WaitGroup

	Logger = log.New(os.Stdout, "", log.LstdFlags)
	Logger.Printf(fmt.Sprintf("%s started\n", APP))

	// configs
	//envSettings := cptplanet.Settings{Prefix: "POSTER_", ErrorOnExtraKeys: false, ErrorOnMissingKeys: true, ErrorOnParseErrors: true}
	//env := cptplanet.NewEnvironment(envSettings)
	//redisServer := env.String("REDIS_SERVER", "127.0.0.1:6379", "host to bind")
	redisServer1 := `127.0.0.1:6379`
	//err := env.Parse()
	//if err != nil {
	//	log.Fatalf("could not get required configs: %v", err)
	//}

	pool := createRedisPool(redisServer1)
	multiplexer := NewMultiplexer()
	multiplexer.SetRedisPool(pool)
	multiplexer.SetPipeWorkerFactory(WorkerFactory)
	multiplexer.SetRejectPipe(createRejectPipe())
	multiplexer.Start()

	Logger.Printf(fmt.Sprintf("%s redis\n", redisServer1))

	scanner := bufio.NewScanner(os.Stdin)

	//TODO: EOF should terminate the program
	for scanner.Scan() {
		m := &PipeMessage{
			payload: scanner.Bytes(),
		}
		m.SetAckHandler(func() error {
			Logger.Printf("Ack called\n")
			wg.Done()
			return nil
		})
		wg.Add(1)

		multiplexer.In() <- m
		Logger.Println("sent multiplex msg")
	}
	if err := scanner.Err(); err != nil {
		Logger.Println(fmt.Sprintf("error reading standard input: %s", err.Error()))
	}

	wg.Wait()
	// need to flush rejected logs because it's buffered
	multiplexer.GetRejectPipe().Stop()
	Logger.Printf(fmt.Sprintf("%s shutting down\n", APP))
}

func createRejectPipe() Pipe {
	rejectLogger, err := syslog.New(syslog.LOG_DEBUG|syslog.LOG_LOCAL1, APP)
	if err != nil {
		log.Fatalf("could not involke syslog logger: %v", err)
	}
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
				fmt.Printf("redis error: %s\n", err.Error())
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
