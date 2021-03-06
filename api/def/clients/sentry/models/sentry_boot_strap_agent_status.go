// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// SentryBootStrapAgentStatus sentry boot strap agent status
//
// swagger:model sentryBootStrapAgentStatus
type SentryBootStrapAgentStatus struct {

	// fingerprint
	Fingerprint string `json:"fingerprint,omitempty"`

	// ip address
	IPAddress string `json:"ipAddress,omitempty"`

	// last checked in
	// Format: date-time
	LastCheckedIn strfmt.DateTime `json:"lastCheckedIn,omitempty"`

	// token state
	TokenState *SentryBootstrapAgentState `json:"tokenState,omitempty"`
}

// Validate validates this sentry boot strap agent status
func (m *SentryBootStrapAgentStatus) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateLastCheckedIn(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTokenState(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SentryBootStrapAgentStatus) validateLastCheckedIn(formats strfmt.Registry) error {
	if swag.IsZero(m.LastCheckedIn) { // not required
		return nil
	}

	if err := validate.FormatOf("lastCheckedIn", "body", "date-time", m.LastCheckedIn.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *SentryBootStrapAgentStatus) validateTokenState(formats strfmt.Registry) error {
	if swag.IsZero(m.TokenState) { // not required
		return nil
	}

	if m.TokenState != nil {
		if err := m.TokenState.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("tokenState")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("tokenState")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this sentry boot strap agent status based on the context it is used
func (m *SentryBootStrapAgentStatus) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateTokenState(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *SentryBootStrapAgentStatus) contextValidateTokenState(ctx context.Context, formats strfmt.Registry) error {

	if m.TokenState != nil {
		if err := m.TokenState.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("tokenState")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("tokenState")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *SentryBootStrapAgentStatus) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *SentryBootStrapAgentStatus) UnmarshalBinary(b []byte) error {
	var res SentryBootStrapAgentStatus
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
