package events

import "uuid"

type GoalEvent struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func NewGoalEvent(
	id uuid.UUID,
	name string,
) *GoalEvent {
	return &GoalEvent{id, name}
}
