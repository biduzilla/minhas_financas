package models

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	Version   int        `db:"version" repo:"version"`
	Deleted   bool       `db:"deleted" repo:"auto"`
	CreatedAt time.Time  `db:"created_at" repo:"auto"`
	CreatedBy *uuid.UUID `db:"created_by" repo:"auto"`
	UpdatedAt *time.Time `db:"updated_at" repo:"auto"`
	UpdatedBy *uuid.UUID `db:"updated_by" repo:"auto"`
}
