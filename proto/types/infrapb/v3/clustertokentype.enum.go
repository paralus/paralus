// Code generated by go generate; DO NOT EDIT.
package infrav3

import (
	bytes "bytes"
	driver "database/sql/driver"
	"fmt"
)

// Scan converts database string to ClusterTokenType
func (e *ClusterTokenType) Scan(value interface{}) error {
	s := value.([]byte)
	*e = ClusterTokenType(ClusterTokenType_value[string(s)])
	return nil
}

// Value converts ClusterTokenType into database string
func (e ClusterTokenType) Value() (driver.Value, error) {
	return ClusterTokenType_name[int32(e)], nil
}

// MarshalJSON converts ClusterTokenType to JSON
func (e ClusterTokenType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("\"")
	buffer.WriteString(e.String())
	buffer.WriteString("\"")
	return buffer.Bytes(), nil
}

// UnmarshalJSON converts ClusterTokenType from JSON
func (e *ClusterTokenType) UnmarshalJSON(b []byte) error {
	if b != nil {
		*e = ClusterTokenType(ClusterTokenType_value[string(b[1:len(b)-1])])
	}
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface
func (e ClusterTokenType) MarshalYAML() (interface{}, error) {
	return ClusterTokenType_name[int32(e)], nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface
func (e *ClusterTokenType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var name string
	if err := unmarshal(&name); err != nil {
		return err
	}

	value, ok := ClusterTokenType_value[name]
	if !ok {
		return fmt.Errorf("invalid ClusterTokenType: %s", name)
	}

	*e = ClusterTokenType(value)
	return nil
}

// implement proto enum interface
func (e ClusterTokenType) IsEnum() {
}
