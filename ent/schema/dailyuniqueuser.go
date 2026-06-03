package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DailyUniqueUser holds one unique user per UTC calendar day.
type DailyUniqueUser struct {
	ent.Schema
}

// Fields of the DailyUniqueUser.
func (DailyUniqueUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("date"),
		field.UUID("user_id", uuid.UUID{}),
	}
}

// Indexes of the DailyUniqueUser.
func (DailyUniqueUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("date", "user_id").Unique(),
	}
}
