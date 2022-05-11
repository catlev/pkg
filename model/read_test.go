package model

import (
	"strings"
	"testing"

	"github.com/catlev/pkg/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	m, err := Read(strings.NewReader(`

entity_type {
	name: "point"

	attribute { name: "x" identifying: "true" type: "integer" }
	attribute { name: "y" type: "integer" }

	relationship { name: "self" impl: "a" }
}

	`))
	require.Nil(t, err)

	assert.Equal(t, &EntityModel{
		Types: Types{{
			Name: "point",
			Attributes: Attributes{
				{Name: "x", Identifying: true, Type: IntegerID},
				{Name: "y", Type: IntegerID},
			},
			Relationships: Relationships{
				{Name: "self", Impl: path.Expr{Kind: path.Rel, Value: "a"}},
			},
		}},
	}, m)
}
