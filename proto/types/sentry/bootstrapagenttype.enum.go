// Code generated by go generate; DO NOT EDIT.
package sentry

import (
	bytes "bytes"
	driver "database/sql/driver"
	"fmt"
)

// Scan converts database string to BootstrapAgentType
func (e *BootstrapAgentType) Scan(value interface{}) error {
	s := value.([]byte)
	*e = BootstrapAgentType(BootstrapAgentType_value[string(s)])
	return nil
}

// Value converts BootstrapAgentType into database string
func (e BootstrapAgentType) Value() (driver.Value, error) {
	return BootstrapAgentType_name[int32(e)], nil
}

// MarshalJSON converts BootstrapAgentType to JSON
func (e BootstrapAgentType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("\"")
	buffer.WriteString(e.String())
	buffer.WriteString("\"")
	return buffer.Bytes(), nil
}

// UnmarshalJSON converts BootstrapAgentType from JSON
func (e *BootstrapAgentType) UnmarshalJSON(b []byte) error {
	if b != nil {
		*e = BootstrapAgentType(BootstrapAgentType_value[string(b[1:len(b)-1])])
	}
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface
func (e BootstrapAgentType) MarshalYAML() (interface{}, error) {
	return BootstrapAgentType_name[int32(e)], nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (e *BootstrapAgentType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var name string
	if err := unmarshal(&name); err != nil {
		return err
	}

	value, ok := BootstrapAgentType_value[name]
	if !ok {
		return fmt.Errorf("invalid BootstrapAgentType: %s", name)
	}

	*e = BootstrapAgentType(value)
	return nil
}

// implement proto enum interface
func (e BootstrapAgentType) IsEnum() {
}
