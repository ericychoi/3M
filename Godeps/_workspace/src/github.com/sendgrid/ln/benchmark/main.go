package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sendgrid/ln"
)

func main() {
	var concurrency int
	var output string
	var statsInterval time.Duration
	var count int

	flag.IntVar(&concurrency, "concurrency", 1000, "number of goroutines writing to ln")
	flag.StringVar(&output, "output", "SYSLOG", "SYSLOG|STDERR")
	flag.DurationVar(&statsInterval, "interval", 5*time.Second, "interval to print stats")
	flag.IntVar(&count, "count", 1000000, "number of messages to log")
	flag.Parse()

	logger := ln.New(output, "DEBUG", "USER", "ln_benchmark")

	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Printf("setting GOMAXPROCS=%d\n", runtime.NumCPU())
	fmt.Printf("sending output to %s\n", output)
	fmt.Printf("starting %d worker goroutines\n", concurrency)
	fmt.Printf("sending %d total messages\n", count)

	// Fill up the buffered channel with work
	c := make(chan int, 1000)
	go func() {
		for i := 0; i < count; i++ {
			c <- i
		}
		close(c)
	}()

	// Keep track of how many messages per second we can log
	messageCount := int64(0)
	start := time.Now()

	// Every (statsInterval), print the stats
	go func() {
		for _ = range time.Tick(statsInterval) {
			elapsed := time.Since(start)
			fmt.Printf("logged %d messages in %s: %0.2f messages/sec\n", atomic.LoadInt64(&messageCount),
				elapsed.String(),
				float64(messageCount)/elapsed.Seconds())

			// Reset the counters
			atomic.StoreInt64(&messageCount, 0)
			start = time.Now()
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for msgNum := range c {
				logger.Info("hello from worker", ln.Map{"id": id, "msg": msgNum})
				atomic.AddInt64(&messageCount, 1)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("done")
}
