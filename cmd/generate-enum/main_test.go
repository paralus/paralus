package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The bulk of this generator's logic is the tmpl constant executed with a
// TemplateData; main() itself is a thin os.Args/os.Exit CLI wrapper around
// it that isn't practical to unit test without a subprocess. This exercises
// the template directly, the same way main() does.
func TestGeneratedEnumTemplate_IsValidGo(t *testing.T) {
	var buf bytes.Buffer
	tp := template.Must(template.New("").Parse(tmpl))
	err := tp.Execute(&buf, TemplateData{PackageName: "v3", EnumName: "ConditionStatus"})
	require.NoError(t, err)

	_, err = parser.ParseFile(token.NewFileSet(), "generated.go", buf.String(), 0)
	require.NoError(t, err, "generated source must be valid Go:\n%s", buf.String())

	out := buf.String()
	assert.Contains(t, out, "package v3")
	assert.Contains(t, out, "func (e *ConditionStatus) Scan(value interface{}) error")
	assert.Contains(t, out, "func (e ConditionStatus) Value() (driver.Value, error)")
	assert.Contains(t, out, "func (e ConditionStatus) MarshalJSON() ([]byte, error)")
	assert.Contains(t, out, "func (e *ConditionStatus) UnmarshalJSON(b []byte) error")
	assert.Contains(t, out, "func (e ConditionStatus) MarshalYAML() (interface{}, error)")
	assert.Contains(t, out, "func (e *ConditionStatus) UnmarshalYAML(unmarshal func(interface{}) error) error")
	assert.Contains(t, out, "func (e ConditionStatus) IsEnum()")
}

func TestGeneratedEnumTemplate_SubstitutesEnumNameEverywhere(t *testing.T) {
	var buf bytes.Buffer
	tp := template.Must(template.New("").Parse(tmpl))
	err := tp.Execute(&buf, TemplateData{PackageName: "userpb", EnumName: "UserType"})
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "{{ .EnumName }}")
	assert.NotContains(t, buf.String(), "{{ .PackageName }}")
}
