package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestValidateEffect(t *testing.T) {
	tests := []struct {
		name    string
		effect  v1.TaintEffect
		wantErr bool
	}{
		{"NoSchedule is valid", v1.TaintEffectNoSchedule, false},
		{"NoExecute is valid", v1.TaintEffectNoExecute, false},
		{"PreferNoSchedule is valid", v1.TaintEffectPreferNoSchedule, false},
		{"unknown effect is invalid", v1.TaintEffect("Bogus"), true},
		{"empty effect is invalid", v1.TaintEffect(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEffect("key", tt.effect)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTaints(t *testing.T) {
	t.Run("all valid taints pass", func(t *testing.T) {
		taints := []*v1.Taint{
			{Key: "key1", Value: "val1", Effect: v1.TaintEffectNoSchedule},
			{Key: "key2", Value: "val2", Effect: v1.TaintEffectNoExecute},
		}
		assert.NoError(t, ValidateTaints(taints))
	})

	t.Run("empty slice is valid", func(t *testing.T) {
		assert.NoError(t, ValidateTaints(nil))
	})

	t.Run("invalid key short-circuits", func(t *testing.T) {
		taints := []*v1.Taint{
			{Key: "", Value: "val1", Effect: v1.TaintEffectNoSchedule},
		}
		assert.Error(t, ValidateTaints(taints))
	})

	t.Run("invalid value short-circuits", func(t *testing.T) {
		taints := []*v1.Taint{
			{Key: "key1", Value: "bad value", Effect: v1.TaintEffectNoSchedule},
		}
		assert.Error(t, ValidateTaints(taints))
	})

	t.Run("invalid effect short-circuits", func(t *testing.T) {
		taints := []*v1.Taint{
			{Key: "key1", Value: "val1", Effect: v1.TaintEffect("Bogus")},
		}
		assert.Error(t, ValidateTaints(taints))
	})
}
