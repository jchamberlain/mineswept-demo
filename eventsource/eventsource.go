package eventsource

import (
	"log"
	"time"

	"github.com/google/uuid"
)

type BaseEvent struct {
	AggregateId string
	Version     int
	At          time.Time
}

func NewAggregateId() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Unable to generate a UUID for a new AggregateId! %s", err)
	}

	return id.String()
}
