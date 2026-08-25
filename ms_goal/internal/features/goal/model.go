package goal

import (
	"fmt"
	"ms_goal/internal/core/domain/models"
	"ms_goal/internal/core/validator"
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
	TargetAmount  int64
	CurrentAmount int64
	Status        GoalStatus
	Deadline      time.Time
	Description   *string
}

type CreateGoalDTO struct {
	Name         string    `json:"name"`
	TargetAmount int64     `json:"target_amount"`
	Deadline     time.Time `json:"deadline"`
	Description  *string   `json:"description,omitempty"`
}

type GoalDTO struct {
	ID            *uuid.UUID `json:"id"`
	UserID        *uuid.UUID `json:"user_id"`
	Name          *string    `json:"name"`
	TargetAmount  *int64     `json:"target_amount"`
	CurrentAmount *int64     `json:"current_amount"`
	Status        *string    `json:"status"`
	Deadline      *time.Time `json:"deadline"`
	Description   *string    `json:"description,omitempty"`
	CreatedAt     *time.Time `json:"created_at"`
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
