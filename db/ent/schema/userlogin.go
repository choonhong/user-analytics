package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserLogin holds the schema definition for the UserLogin entity.
type UserLogin struct {
	ent.Schema
}

// Fields of the UserLogin.
func (UserLogin) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Time("login_time"),
	}
}

// Indexes of the UserLogin.
func (UserLogin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "login_time").Unique(),
		index.Fields("login_time"),
	}
}
