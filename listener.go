package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/lib/pq"
)

func worker(newEventSignals chan int, tickDuration time.Duration, callback func() bool) {
	queue := make([]int, 0)

	requestEvent := func() {
		if len(queue) > 0 {
			if callback() {
				queue = queue[1:]
			}
		}
	}

	for {
		select {
		case <-time.After(tickDuration):
			requestEvent()
			continue
		case <-newEventSignals:
			queue = append(queue, 1)
			requestEvent()
			continue
		}
	}
}

func createListener(wg *sync.WaitGroup, dbURL string, s subscriber) error {
	sigquit := make(chan os.Signal, 1)
	signal.Notify(sigquit, os.Interrupt, os.Kill)
	log.Printf("[%d] Listening...\n", s.ID)

	db, err := sql.Open("postgres", dbURL)
	defer db.Close()

	listener := pq.NewListener(dbURL, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Println(err.Error())
		}
	})

	err = listener.Listen(s.EventName)
	if err != nil {
		log.Fatalf("[%d] Can't listen to %s\n", s.ID, s.EventName)
	}

	newEventSignals := make(chan int)

	go worker(newEventSignals, s.TickDuration, func() bool {
		return s.sendEvent(db)
	})

	for {
		select {
		case <-listener.Notify:
			log.Printf("[%d] RECEIVED\n", s.ID)
			newEventSignals <- 1
			continue
		case <-time.After(90 * time.Second):
			log.Printf("[%d] Ping...\n", s.ID)
			listener.Ping()
			continue
		case <-sigquit:
			log.Printf("[%d] Gracefully shutting down: %s...\n", s.ID, s.URL)
			wg.Done()
			break
		}
	}
}
