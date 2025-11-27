package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Log struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Domain    string    `json:"domain" db:"domain"`
	Action    string    `json:"action" db:"action"`
	Content   Content   `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Content map[string]interface{}

func (c *Content) Scan(value interface{}) error {
	if value == nil {
		*c = make(Content)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		*c = make(Content)
		return nil
	}
}

func (c Content) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

type InitializeRequest struct {
	RecordCount int    `json:"record_count" binding:"required,oneof=1000 10000 100000 1000000 10000000"`
	ContentSize string `json:"content_size" binding:"required,oneof=small medium large"`
}

type LogFilter struct {
	UserID      *string `form:"user_id"`
	Domain      *string `form:"domain"`
	CreatedAt   *string `form:"created_at"`
	CreatedAtTo *string `form:"created_at_to"`
	ContentLike *string `form:"content_like"`
	SearchTerm  *string `form:"search_term"`
	Page        int     `form:"page,default=1"`
	Limit       int     `form:"limit,default=50"`
}
