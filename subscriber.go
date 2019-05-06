package main

import (
	"database/sql"
	"gossipy/web"
	"log"
	"time"
)

type subscriber struct {
	ID             int64
	URL            string
	Method         string
	EventName      string
	AcceptedStatus int
	TickDuration   time.Duration
}

type event struct {
	eventID      int64
	notification string
}

const lastEventQuery = `SELECT
		id,
		notification
	FROM
		events e
	INNER JOIN
		event_subscriptions es
	ON
		es.event_id = e.id
	WHERE
		es.subscriber_id = $1
	ORDER BY e.id ASC
	LIMIT 1
`

const deleteSubscriberEventQuery = `
	DELETE FROM
		event_subscriptions es
	WHERE
		es.event_id = $1 AND es.subscriber_id = $2
`

func (s *subscriber) sendEvent(db *sql.DB) bool {
	var eventID int64
	var notification string

	row := db.QueryRow(lastEventQuery, s.ID)
	row.Scan(
		&eventID,
		&notification,
	)

	if eventID == 0 {
		return false
	}

	log.Println("Received:", eventID, notification)

	statusCode, err := web.Hook(s.URL, s.Method, &notification)

	if err != nil {
		log.Printf("[%d] Couldn't deliver %s\n", s.ID, err)
		return false
	}

	if !s.validateStatusCode(statusCode) {
		log.Printf("[%d] Status code %d not accepted, expected: %d\n", s.ID, statusCode, s.AcceptedStatus)
		return false
	}

	log.Printf("[%d] %s %s\n", s.ID, s.Method, s.URL)

	_, err = db.Exec(deleteSubscriberEventQuery, eventID, s.ID)
	if err != nil {
		return false
	}

	log.Printf("[%d] Removed from queue\n", s.ID)
	return true
}

func (s *subscriber) validateStatusCode(statusCode int) bool {
	if s.AcceptedStatus == -1 {
		return true
	}

	return statusCode == s.AcceptedStatus
}
