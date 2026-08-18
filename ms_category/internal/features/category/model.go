package category

import (
	"fmt"
	"ms_category/internal/core/domain/models"
	"ms_category/internal/core/validator"

	"github.com/google/uuid"
)

type CategoryType int

const (
	CategoryTypeInput CategoryType = iota
	CategoryTypeOutput
)

type Category struct {
	models.BaseModel
	ID     uuid.UUID
	UserID uuid.UUID
	Name   string
	Type   CategoryType
	GoalID *uuid.UUID
}

type CategoryDTO struct {
	ID     *uuid.UUID `json:"id,omitempty"`
	UserID *uuid.UUID `json:"user_id,omitempty"`
	Name   *string    `json:"name,omitempty"`
	Type   *string    `json:"type"`
	GoalID *uuid.UUID `json:"goal_id,omitempty"`
}

func (d CategoryDTO) ToModel() *Category {
	var model Category

	if d.ID != nil {
		model.ID = *d.ID
	}
	if d.UserID != nil {
		model.UserID = *d.UserID
	}
	if d.Name != nil {
		model.Name = *d.Name
	}
	if d.GoalID != nil {
		model.GoalID = d.GoalID
	}

	if d.Type != nil {
		if ct, err := ParseCategoryType(*d.Type); err == nil {
			model.Type = ct
		}
	}

	return &model
}

func (m *Category) ToDTO() CategoryDTO {
	t := m.Type.String()

	return CategoryDTO{
		ID:     &m.ID,
		UserID: &m.UserID,
		Name:   &m.Name,
		Type:   &t,
		GoalID: m.GoalID,
	}
}

func (ct CategoryType) String() string {
	switch ct {
	case CategoryTypeInput:
		return "input"
	case CategoryTypeOutput:
		return "output"
	default:
		return "unknow"
	}
}

func ParseCategoryType(s string) (CategoryType, error) {
	switch s {
	case "input":
		return CategoryTypeInput, nil
	case "output":
		return CategoryTypeOutput, nil
	default:
		return 0, fmt.Errorf("invalid category type: %s", s)
	}
}

func (m *Category) Validate(v *validator.Validator) {
	v.Check(m.Name != "", "name", "must be provided")
	v.Check(len(m.Name) >= 3, "name", "must be at least 3 characters long")
	v.Check(len(m.Name) <= 100, "name", "must not be more than 100 characters long")

	v.Check(m.Type == CategoryTypeInput || m.Type == CategoryTypeOutput, "type", "must be a valid category type")
}
