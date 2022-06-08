package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/paralus/paralus/internal/cluster/constants"
)

func IsValidKubernetesLabelNameValueRegex(input string) bool {
	allSubStrings := constants.KubernetesLabelNameRegex.FindAllString(input, -1)
	return len(allSubStrings) == 1 && allSubStrings[0] == input
}

func extractCustomLabelKeyParts(labelKey string) (string, string) {
	prefixAndKey := strings.SplitN(labelKey, "/", 2)
	if len(prefixAndKey) == 1 {
		return "", prefixAndKey[0]
	} else {
		return prefixAndKey[0], prefixAndKey[1]
	}
}

func ValidateCustomLabelKey(k, v string) error {
	if k == "" {
		return fmt.Errorf("invalid custom label key; key shouldn't be empty, but received an empty key for value: %s", v)
	}
	prefix, suffix := extractCustomLabelKeyParts(k)
	err := validateCustomLabelKeyPrefix(prefix)
	if err != nil {
		return err
	}
	err = validateCustomLabelKeySuffix(suffix, v)
	return err
}

func validateCustomLabelKeyPrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	if len(prefix) > 253 {
		return fmt.Errorf("invalid custom label key; label key prefix should be less than or equal to 253 characters (%s)", prefix)
	}
	if prefix == constants.ParalusDomainLabel {
		return fmt.Errorf("invalid custom label key; paralus.dev is a reserved domain for Paralus and custom labels shouldn't use this domain: %s", prefix)
	}
	if !IsValidKubernetesLabelNameValueRegex(prefix) {
		return fmt.Errorf("invalid custom label key prefix (%s); prefix should start and end with alpha numerical value and can have dashes (-), underscores (_) and dots (.)", prefix)
	}
	return nil
}

func validateCustomLabelKeySuffix(key, value string) error {
	if key == "" {
		return fmt.Errorf("invalid custom label key; key shouldn't be empty, but received an empty key for value: %s", value)
	}
	if len(key) > 63 {
		return fmt.Errorf("invalid custom label key; label keys should be less than or equal to 63 characters (%s)", key)
	}
	if !IsValidKubernetesLabelNameValueRegex(key) {
		return fmt.Errorf("invalid custom label key (%s); key should start and end with alpha numerical value and can have dashes (-), underscores (_) and dots (.)", key)
	}
	return nil
}

func validateCustomLabelValue(key, value string) error {
	if value == "" {
		return nil
	}
	if len(value) > 63 {
		return fmt.Errorf("invalid custom label value; label keys should be less than or equal to 63 characters (%s)", key)
	}
	if !IsValidKubernetesLabelNameValueRegex(value) {
		return fmt.Errorf("invalid custom label value (%s) for key (%s); value should start and end with alpha numerical value and can have dashes (-), underscores (_) and dots (.)", value, key)
	}
	return nil
}

func ValidateCustomLabels(labels map[string]string) error {
	for k, v := range labels {
		err := ValidateCustomLabelKey(k, v)
		if err != nil {
			return err
		}
		err = validateCustomLabelValue(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func SanitizeLabelValues(labelValue string) string {
	// Remove all extra spaces between words
	// replace spaces and / to -

	//For now just make the value lower cased and replace / with -; add more as we discover conditions
	labelValue = strings.TrimSpace(labelValue)
	space := regexp.MustCompile(`\s+`)
	labelValue = space.ReplaceAllString(labelValue, "-")
	sanitizeLabelValue := strings.ReplaceAll(strings.ToLower(labelValue), "/", "-")

	return sanitizeLabelValue
}

func ValidateFullKey(k, v string) error {
	if k == "" {
		return fmt.Errorf("invalid custom label key; key shouldn't be empty, but received an empty key for value: %s", v)
	}
	prefix, key := getLabelPrefixAndKey(k)
	err := ValidatePrefix(prefix)
	if err != nil {
		return err
	}
	err = ValidateKey(key, v)
	return err
}

func getLabelPrefixAndKey(labelKey string) (string, string) {
	prefixAndKey := strings.SplitN(labelKey, "/", 2)
	if len(prefixAndKey) == 1 {
		return "", prefixAndKey[0]
	} else {
		return prefixAndKey[0], prefixAndKey[1]
	}
}

func ValidateKey(key, value string) error {
	if key == "" {
		return fmt.Errorf("invalid custom label key; key shouldn't be empty, but received an empty key for value: %s", value)
	}
	if len(key) > 63 {
		return fmt.Errorf("invalid custom label key; label keys should be less than or equal to 63 characters (%s)", key)
	}
	if !IsValidKubernetesLabelNameValueRegex(key) {
		return fmt.Errorf("invalid custom label key (%s); key should start and end with alpha numerical value and can have dashes (-), underscores (_) and dots (.)", key)
	}
	return nil
}

func ValidatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	if len(prefix) > 253 {
		return fmt.Errorf("invalid custom label key; label key prefix should be less than or equal to 253 characters (%s)", prefix)
	}
	if prefix == constants.ParalusDomainLabel {
		return fmt.Errorf("invalid custom label key; paralus.dev is a reserved domain for Paralus and custom labels shouldn't use this domain: %s", prefix)
	}
	if !IsValidKubernetesLabelNameValueRegex(prefix) {
		return fmt.Errorf("invalid custom label key prefix (%s); prefix should start and end with alpha numerical value and can have dashes (-), underscores (_) and dots (.)", prefix)
	}
	return nil
}

func ValidateValue(key, value string) error {
	if value == "" {
		return nil
	}
	if len(value) > 63 {
		return fmt.Errorf("invalid custom label value; label keys should be less than or equal to 63 characters (%s)", key)
	}
	if !IsValidKubernetesLabelNameValueRegex(value) {
		return fmt.Errorf("invalid custom label value (%s) for key (%s); value should start and end with alpha numerical value and can have dashes (-), underscores (_) and dots (.)", value, key)
	}
	return nil
}
