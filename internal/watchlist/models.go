package watchlist

import (
	"time"

	"github.com/google/uuid"
)

type UUID = uuid.UUID

func ParseUUID(s string) (UUID, error) {
	return uuid.Parse(s)
}

type WatchedTrain struct {
	ID          uuid.UUID `json:"id"`
	TrainNumber int       `json:"trainNumber"`
	OriginID    string    `json:"originId"`
	OriginName  string    `json:"originName"`
	Destination string    `json:"destination"`
	DaysOfWeek  string    `json:"daysOfWeek,omitempty"`
	Notes       string    `json:"notes,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TrainCheck struct {
	ID           uuid.UUID `json:"id"`
	WatchedID    uuid.UUID `json:"watchedId"`
	TrainNumber  int       `json:"trainNumber"`
	Delay        int       `json:"delay"`
	Status       string    `json:"status"`
	CheckedAt    time.Time `json:"checkedAt"`
}
