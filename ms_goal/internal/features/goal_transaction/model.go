package goaltransaction

import (
	"ms_goal/internal/core/domain/models"
	"ms_goal/internal/features/goal"
	"uuid"
)

type GoalDetails struct {
	models.BaseModel
	ID            uuid.UUID
	Goal          goal.Goal
	TransactionID uuid.UUID
	Amount        float64
}
