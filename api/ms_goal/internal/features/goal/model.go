package goal

import (
	"fmt"
	"shared/auth/domain/models"
	"shared/validator"
	"time"
	"uuid"
)

type GoalStatus int

const (
	GoalStatusInProgress GoalStatus = iota
	GoalStatusCompleted
	GoalStatusExpired
	GoalStatusCanceled
)

func (s GoalStatus) String() string {
	switch s {
	case GoalStatusInProgress:
		return "IN_PROGRESS"
	case GoalStatusCompleted:
		return "COMPLETED"
	case GoalStatusExpired:
		return "EXPIRED"
	case GoalStatusCanceled:
		return "CANCELED"
	default:
		return "unknow"
	}
}

func ParseGoalStatus(s string) (GoalStatus, error) {
	switch s {
	case "em andamento":
		return GoalStatusInProgress, nil
	case "concluído":
		return GoalStatusCompleted, nil
	case "vencido":
		return GoalStatusExpired, nil
	case "cancelado":
		return GoalStatusCanceled, nil
	default:
		return 0, fmt.Errorf("invalid goal status: %s", s)
	}
}

type Goal struct {
	models.BaseModel
	ID            uuid.UUID
	UserID        uuid.UUID
	Name          string
	TargetAmount  float64
	CurrentAmount float64
	Status        GoalStatus
	Deadline      time.Time
	Description   *string
}

type CreateGoalDTO struct {
	Name         string    `json:"name"`
	TargetAmount float64   `json:"target_amount"`
	Deadline     time.Time `json:"deadline"`
	Description  *string   `json:"description,omitempty"`
}

type GoalDTO struct {
	ID            *uuid.UUID `json:"id"`
	UserID        *uuid.UUID `json:"user_id"`
	Name          *string    `json:"name"`
	TargetAmount  *float64   `json:"target_amount"`
	CurrentAmount *float64   `json:"current_amount"`
	Status        *string    `json:"status"`
	Deadline      *time.Time `json:"deadline"`
	Description   *string    `json:"description,omitempty"`
	CreatedAt     *time.Time `json:"created_at"`
	Version       *int       `json:"version"`
}

type GoalReportDTO struct {
	Goal              *GoalDTO   `json:"goal"`
	TotalContributed  float64    `json:"total_contributed"`
	Progress          float64    `json:"progress"`
	ValuePerMonth     float64    `json:"value_per_month"`
	RemainingAmount   float64    `json:"remaining_amount"`
	TransactionsCount int        `json:"transactions_count"`
	LastContribution  *time.Time `json:"last_contribution,omitempty"`
}

func (d CreateGoalDTO) ToModel() *Goal {
	return &Goal{
		Name:         d.Name,
		TargetAmount: d.TargetAmount,
		Deadline:     d.Deadline,
		Description:  d.Description,
		Status:       GoalStatusInProgress,
	}
}

func (d GoalDTO) ToModel() *Goal {
	var model Goal

	if d.ID != nil {
		model.ID = *d.ID
	}
	if d.UserID != nil {
		model.UserID = *d.UserID
	}
	if d.Name != nil {
		model.Name = *d.Name
	}
	if d.TargetAmount != nil {
		model.TargetAmount = *d.TargetAmount
	}
	if d.CurrentAmount != nil {
		model.CurrentAmount = *d.CurrentAmount
	}
	if d.Description != nil {
		model.Description = d.Description
	}
	if d.Deadline != nil {
		model.Deadline = *d.Deadline
	}

	if d.Version != nil {
		model.Version = *d.Version
	}

	if d.Status != nil {
		if st, err := ParseGoalStatus(*d.Status); err == nil {
			model.Status = st
		}
	}

	return &model
}

func (m *Goal) ToDTO() GoalDTO {
	statusStr := m.Status.String()
	return GoalDTO{
		ID:            &m.ID,
		UserID:        &m.UserID,
		Name:          &m.Name,
		TargetAmount:  &m.TargetAmount,
		CurrentAmount: &m.CurrentAmount,
		Status:        &statusStr,
		Deadline:      &m.Deadline,
		Description:   m.Description,
		CreatedAt:     &m.CreatedAt,
		Version:       &m.Version,
	}
}

func (m *Goal) Validate(v *validator.Validator) {
	v.Check(m.Name != "", "name", "must be provided")
	v.Check(len(m.Name) <= 100, "name", "must not be more than 100 characters long")

	v.Check(m.TargetAmount > 0, "target_amount", "must be greater than zero")
	v.Check(m.CurrentAmount >= 0, "current_amount", "must not be negative")

	if m.Deadline.IsZero() {
		v.AddError("deadline", "must be provided")
	}

	if m.Status < GoalStatusInProgress || m.Status > GoalStatusCanceled {
		v.AddError("status", "must be a valid status")
	}
}

func (g *Goal) Progress() float64 {
	if g.TargetAmount <= 0 {
		return 0
	}

	p := float64(g.CurrentAmount) / float64(g.TargetAmount) * 100
	if p > 100 {
		return 100
	}
	return p
}

func (g *Goal) RemainingAmount() float64 {
	r := g.TargetAmount - g.CurrentAmount
	if r < 0 {
		return 0
	}

	return float64(r)
}

func (g *Goal) SuggestedMonthly(now time.Time) float64 {
	if !g.Deadline.After(now) {
		return 0
	}

	months := 0
	cursor := now
	for cursor.Before(g.Deadline) {
		cursor = cursor.AddDate(0, 1, 0)
		months++
	}

	if months < 1 {
		months = 1
	}

	return g.RemainingAmount() / float64(months)
}
