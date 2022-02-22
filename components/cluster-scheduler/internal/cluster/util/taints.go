package util

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

func ValidateTaints(taints []*v1.Taint) error {
	for _, taint := range taints {
		err := ValidateFullKey(taint.Key, taint.Value)
		if err != nil {
			return err
		}
		err = ValidateValue(taint.Key, taint.Value)
		if err != nil {
			return err
		}
		err = ValidateEffect(taint.Key, taint.Effect)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateEffect(key string, effect v1.TaintEffect) error {
	switch effect {
	case v1.TaintEffectNoSchedule, v1.TaintEffectNoExecute, v1.TaintEffectPreferNoSchedule:
		return nil
	default:
		return fmt.Errorf("invalid taint effect '%s' for the key '%s'", effect, key)
	}
}
