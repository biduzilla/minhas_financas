package events

import "uuid"

type TransactionEvent struct {
	ID     uuid.UUID `json:"id"`
	Amount float64   `json:"name"`
	UserID uuid.UUID `json:"user_id"`
	GoalID uuid.UUID `json:"goal_id"`
}

func NewTransactionEvent(
	id uuid.UUID,
	amount float64,
	userID uuid.UUID,
	goalID uuid.UUID,
) TransactionEvent {
	return TransactionEvent{id, amount, userID, goalID}
}
