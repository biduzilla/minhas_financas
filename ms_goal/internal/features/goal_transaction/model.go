package goaltransaction

import (
	"ms_goal/internal/core/domain/models"
	"ms_goal/internal/core/validator"
	"time"
	"uuid"
)

type GoalTransaction struct {
	models.BaseModel
	ID            uuid.UUID
	UserID        uuid.UUID
	GoalID        uuid.UUID
	TransactionID uuid.UUID
	Amount        float64
}

type CreateGoalTransactionDTO struct {
	GoalID        uuid.UUID `json:"goal_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	Amount        float64   `json:"amount"`
}

type GoalTransactionDTO struct {
	ID            *uuid.UUID `json:"id,omitempty"`
	GoalID        *uuid.UUID `json:"goal_id,omitempty"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty"`
	Amount        *float64   `json:"amount,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

func (d CreateGoalTransactionDTO) ToModel() *GoalTransaction {
	return &GoalTransaction{
		GoalID:        d.GoalID,
		TransactionID: d.TransactionID,
		Amount:        d.Amount,
	}
}

func (d GoalTransactionDTO) ToModel() *GoalTransaction {
	var model GoalTransaction

	if d.ID != nil {
		model.ID = *d.ID
	}
	if d.GoalID != nil {
		model.GoalID = *d.GoalID
	}
	if d.TransactionID != nil {
		model.TransactionID = *d.TransactionID
	}
	if d.Amount != nil {
		model.Amount = *d.Amount
	}
	if d.CreatedAt != nil {
		model.CreatedAt = *d.CreatedAt
	}
	return &model
}

func (m *GoalTransaction) ToDTO() GoalTransactionDTO {
	return GoalTransactionDTO{
		ID:            &m.ID,
		GoalID:        &m.GoalID,
		TransactionID: &m.TransactionID,
		Amount:        &m.Amount,
		CreatedAt:     &m.CreatedAt,
	}
}

func (m *GoalTransaction) Validate(v *validator.Validator) {
	v.Check(m.GoalID != uuid.Nil(), "goal_id", "must be provided")
	v.Check(m.TransactionID != uuid.Nil(), "transaction_id", "must be provided")
	v.Check(m.Amount > 0, "amount", "must be greater than zero")
}
