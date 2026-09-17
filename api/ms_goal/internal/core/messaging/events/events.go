package events

import "uuid"

const (
	TransactionGoalCreatedTopic = "transaction_goal_created"
	TransactionGoalDeletedTopic = "transaction_goal_deleted"
)

type GoalEvent struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	UserID uuid.UUID `json:"user_id"`
}

type TransactionEvent struct {
	ID     uuid.UUID `json:"id"`
	Amount float64   `json:"name"`
	UserID uuid.UUID `json:"user_id"`
	GoalID uuid.UUID `json:"goal_id"`
}

func (e TransactionEvent) Key() string {
	return e.ID.String()
}

func (e GoalEvent) Key() string {
	return e.ID.String()
}

func NewGoalEvent(
	id uuid.UUID,
	name string,
	userID uuid.UUID,
) GoalEvent {
	return GoalEvent{id, name, userID}
}
