// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// RPCRegisterAgentResponse rpc register agent response
//
// swagger:model rpcRegisterAgentResponse
type RPCRegisterAgentResponse struct {

	// ca certificate
	// Format: byte
	CaCertificate strfmt.Base64 `json:"caCertificate,omitempty"`

	// certificate
	// Format: byte
	Certificate strfmt.Base64 `json:"certificate,omitempty"`
}

// Validate validates this rpc register agent response
func (m *RPCRegisterAgentResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this rpc register agent response based on context it is used
func (m *RPCRegisterAgentResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *RPCRegisterAgentResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RPCRegisterAgentResponse) UnmarshalBinary(b []byte) error {
	var res RPCRegisterAgentResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
