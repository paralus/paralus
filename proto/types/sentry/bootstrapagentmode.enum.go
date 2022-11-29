// Code generated by go generate; DO NOT EDIT.
package sentry

import (
	bytes "bytes"
	driver "database/sql/driver"
)

// Scan converts database string to BootstrapAgentMode
func (e *BootstrapAgentMode) Scan(value interface{}) error {
	s := value.([]byte)
	*e = BootstrapAgentMode(BootstrapAgentMode_value[string(s)])
	return nil
}

// Value converts BootstrapAgentMode into database string
func (e BootstrapAgentMode) Value() (driver.Value, error) {
	return BootstrapAgentMode_name[int32(e)], nil
}

// MarshalJSON converts BootstrapAgentMode to JSON
func (e BootstrapAgentMode) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("\"")
	buffer.WriteString(e.String())
	buffer.WriteString("\"")
	return buffer.Bytes(), nil
}

// UnmarshalJSON converts BootstrapAgentMode from JSON
func (e *BootstrapAgentMode) UnmarshalJSON(b []byte) error {
	if b != nil {
		*e = BootstrapAgentMode(BootstrapAgentMode_value[string(b[1:len(b)-1])])
	}
	return nil
}

// implement proto enum interface
func (e BootstrapAgentMode) IsEnum() {
}