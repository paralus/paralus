package enforcer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestKeyMatchCu(t *testing.T) {
	tests := []struct {
		name string
		key1 string
		key2 string
		want bool
	}{
		{
			name: "key2 star wildcard always matches, even empty key1",
			key1: "",
			key2: "*",
			want: true,
		},
		{
			name: "key2 star wildcard matches any key1",
			key1: "admin:ops1",
			key2: "*",
			want: true,
		},
		{
			name: "exact literal match",
			key1: "abc",
			key2: "abc",
			want: true,
		},
		{
			name: "literal mismatch",
			key1: "abc",
			key2: "xyz",
			want: false,
		},
		{
			name: "path glob suffix matches single segment",
			key1: "/api/v1/foo",
			key2: "/api/v1/*",
			want: true,
		},
		{
			name: "path glob suffix does not match shorter prefix",
			key1: "/api/v1",
			key2: "/api/v1/*",
			want: false,
		},
		{
			name: "colon param matches one path segment",
			key1: "/api/v1/foo",
			key2: "/api/v1/:id",
			want: true,
		},
		{
			name: "colon param does not span multiple path segments",
			key1: "/api/v1/foo/bar",
			key2: "/api/v1/:id",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, KeyMatchCu(tt.key1, tt.key2))
		})
	}
}

func TestNewCasbinEnforcer(t *testing.T) {
	db := &gorm.DB{}

	e := NewCasbinEnforcer(db)

	assert.NotNil(t, e)
	assert.Same(t, db, e.db)
}
