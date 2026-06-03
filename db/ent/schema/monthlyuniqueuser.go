package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MonthlyUniqueUser holds one unique user per UTC calendar month.
type MonthlyUniqueUser struct {
	ent.Schema
}

// Fields of the MonthlyUniqueUser.
func (MonthlyUniqueUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("month"),
		field.UUID("user_id", uuid.UUID{}),
	}
}

// Indexes of the MonthlyUniqueUser.
func (MonthlyUniqueUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("month", "user_id").Unique(),
	}
}
