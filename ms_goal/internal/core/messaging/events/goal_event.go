package events

import "uuid"

type GoalEvent struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	UserID uuid.UUID `json:"user_id"`
}

func NewGoalEvent(
	id uuid.UUID,
	name string,
	userID uuid.UUID,
) GoalEvent {
	return GoalEvent{id, name, userID}
}
