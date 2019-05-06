package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var defaultDbURL = "postgres://postgres:postgres@localhost:5432/dictionary?sslmode=disable"
var defaultLogPath = filepath.Join("log", "gossip.log")

func main() {

	dbURL := flag.String("db", defaultDbURL, "database url")
	logPath := flag.String("l", defaultLogPath, "log path")

	flag.Parse()

	f, err := os.OpenFile(*logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

	if os.IsNotExist(err) {
		log.Fatalf("log directory does not exists, %v", err)
	}
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println("Starting server...")

	subscribers := []subscriber{
		subscriber{
			ID:             1,
			EventName:      "words",
			URL:            "http://localhost:4000/receive_notification/a",
			AcceptedStatus: 200,
			TickDuration:   10 * time.Millisecond,
			Method:         "POST",
		},
		subscriber{
			ID:             2,
			EventName:      "words",
			URL:            "http://localhost:4000/receive_notification/b",
			AcceptedStatus: 200,
			TickDuration:   10 * time.Millisecond,
			Method:         "POST",
		},
	}

	var wg sync.WaitGroup
	wg.Add(len(subscribers))

	for _, subscriber := range subscribers {
		go createListener(&wg, *dbURL, subscriber)
	}

	wg.Wait()
	log.Println("finished")
}
