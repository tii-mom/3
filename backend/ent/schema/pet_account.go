package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PetAccount holds the schema definition for the PetAccount entity.
// Represents an AI pet from TAI Protocol that has been provisioned with
// a 3api account and API key for compute access.
type PetAccount struct {
	ent.Schema
}

func (PetAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "pet_accounts"},
	}
}

func (PetAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (PetAccount) Fields() []ent.Field {
	return []ent.Field{
		// TAI Protocol pet identifier (e.g. "gen0-001")
		field.String("pet_id").
			MaxLen(64).
			NotEmpty().
			Unique().
			Comment("External pet ID from TAI Protocol"),

		// The 3api user account created for this pet
		field.Int64("user_id").
			Comment("3api user account owning this pet's API key"),

		// The API key provisioned for the pet
		field.Int64("api_key_id").
			Optional().
			Nillable().
			Comment("Provisioned API key ID"),

		// Group assignment for pricing tier
		field.Int64("group_id").
			Optional().
			Nillable().
			Comment("Pricing group (tai-pets bulk tier)"),

		// Owner's Telegram user ID (for attribution)
		field.String("owner_tg_id").
			MaxLen(64).
			Optional().
			Comment("Pet owner Telegram user ID"),

		// TAI token tracking
		field.Float("tai_spent_total").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("Total TAI tokens spent on compute"),

		field.Float("compute_credits_total").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("Total compute credits granted (3api balance)"),

		// Status: active, suspended, deleted
		field.String("status").
			MaxLen(20).
			Default("active"),

		// Rate limiting for pets (prevent abuse)
		field.Float("daily_tai_limit").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(100).
			Comment("Max TAI spend per day"),

		field.Float("daily_tai_used").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),

		field.Time("daily_reset_at").
			Optional().
			Nillable(),
	}
}

func (PetAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("api_key", APIKey.Type).
			Field("api_key_id").
			Unique(),
	}
}

func (PetAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("pet_id").Unique(),
		index.Fields("user_id"),
		index.Fields("owner_tg_id"),
		index.Fields("status"),
	}
}
