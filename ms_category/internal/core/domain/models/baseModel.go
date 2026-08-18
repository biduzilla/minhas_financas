package models

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	Version   int        `db:"version"`
	Deleted   bool       `db:"deleted"`
	CreatedAt time.Time  `db:"created_at"`
	CreatedBy *uuid.UUID `db:"created_by"`
	UpdatedAt *time.Time `db:"updated_at"`
	UpdatedBy *uuid.UUID `db:"updated_by" `
}
