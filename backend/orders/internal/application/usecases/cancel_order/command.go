package cancel_order

import "github.com/google/uuid"

type Command struct {
	OrderGroupID uuid.UUID
	TraderID     uuid.UUID
}

type Result struct {
	OrderGroupID string
	Status       string
}
