// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// RPCLookupClusterResponse rpc lookup cluster response
//
// swagger:model rpcLookupClusterResponse
type RPCLookupClusterResponse struct {

	// display name
	DisplayName string `json:"displayName,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

// Validate validates this rpc lookup cluster response
func (m *RPCLookupClusterResponse) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this rpc lookup cluster response based on context it is used
func (m *RPCLookupClusterResponse) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *RPCLookupClusterResponse) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RPCLookupClusterResponse) UnmarshalBinary(b []byte) error {
	var res RPCLookupClusterResponse
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
