package audit

import (
	"encoding/json"

	"creaciones-api/internal/db"
)

type Actor struct {
	UserID   *int
	Username *string
}

type Event struct {
	ID            int     `json:"id"`
	EntityType    string  `json:"entity_type"`
	EntityID      int     `json:"entity_id"`
	Action        string  `json:"action"`
	ActorUserID   *int    `json:"actor_user_id"`
	ActorUsername *string `json:"actor_username"`
	Summary       *string `json:"summary"`
	BeforeJSON    *string `json:"before_json"`
	AfterJSON     *string `json:"after_json"`
	CreatedAt     db.Time `json:"created_at"`
}

type CreateEventDTO struct {
	EntityType string
	EntityID   int
	Action     string
	Actor      Actor
	Summary    string
	Before     any
	After      any
}

type FilterDTO struct {
	EntityType *string
	EntityID   *int
	From       *string
	To         *string
}

func marshalSnapshot(value any) (*string, error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	text := string(data)
	return &text, nil
}
