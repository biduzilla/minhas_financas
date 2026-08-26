package transactions

import (
	"fmt"
	"ms_transaction/internal/core/domain/models"
	"ms_transaction/internal/core/validator"
	"time"
	"uuid"
)

type Transaction struct {
	models.BaseModel
	ID          uuid.UUID
	UserID      uuid.UUID
	Amount      float64
	CategoryID  uuid.UUID
	Description string
}

type TransactionDTO struct {
	models.BaseModel
	ID          *uuid.UUID `json:"id"`
	Amount      *float64   `json:"amount"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Description *string    `json:"description"`
	Version     *int       `json:"version"`
	CreatedAt   *time.Time `json:"created_at"`
}

func (m *Transaction) ToDTO() TransactionDTO {
	return TransactionDTO{
		ID:          &m.ID,
		Amount:      &m.Amount,
		CategoryID:  &m.CategoryID,
		Description: &m.Description,
		Version:     &m.Version,
		CreatedAt:   &m.CreatedAt,
	}
}

func (d TransactionDTO) ToModel() *Transaction {
	var model Transaction

	if d.ID != nil {
		model.ID = *d.ID
	}
	if d.CategoryID != nil {
		model.CategoryID = *d.CategoryID
	}
	if d.Amount != nil {
		model.Amount = *d.Amount
	}
	if d.Description != nil {
		model.Description = *d.Description
	}
	if d.Version != nil {
		model.Version = *d.Version
	}

	return &model
}

func (m *Transaction) Validate(v *validator.Validator) {
	v.Check(m.Amount > 0, "amount", "must be provided and positive")
	v.Check(m.CategoryID != uuid.Nil(), "category_id", "must be provided")
	v.Check(len(m.Description) <= 100, "description", "must not be more than 100 characters long")
}

type categoryType int

const (
	categoryTypeInput categoryType = iota
	categoryTypeOutput
)

func parseCategoryType(s string) (categoryType, error) {
	switch s {
	case "input":
		return categoryTypeInput, nil
	case "output":
		return categoryTypeOutput, nil
	default:
		return 0, fmt.Errorf("invalid category type: %s", s)
	}
}
