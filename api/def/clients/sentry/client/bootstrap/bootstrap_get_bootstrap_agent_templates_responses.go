// Code generated by go-swagger; DO NOT EDIT.

package bootstrap

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/paralus/paralus/api/def/clients/sentry/models"
)

// BootstrapGetBootstrapAgentTemplatesReader is a Reader for the BootstrapGetBootstrapAgentTemplates structure.
type BootstrapGetBootstrapAgentTemplatesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *BootstrapGetBootstrapAgentTemplatesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewBootstrapGetBootstrapAgentTemplatesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 403:
		result := NewBootstrapGetBootstrapAgentTemplatesForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewBootstrapGetBootstrapAgentTemplatesNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewBootstrapGetBootstrapAgentTemplatesInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewBootstrapGetBootstrapAgentTemplatesDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewBootstrapGetBootstrapAgentTemplatesOK creates a BootstrapGetBootstrapAgentTemplatesOK with default headers values
func NewBootstrapGetBootstrapAgentTemplatesOK() *BootstrapGetBootstrapAgentTemplatesOK {
	return &BootstrapGetBootstrapAgentTemplatesOK{}
}

/*
	BootstrapGetBootstrapAgentTemplatesOK describes a response with status code 200, with default header values.

A successful response.
*/
type BootstrapGetBootstrapAgentTemplatesOK struct {
	Payload *models.SentryBootstrapAgentTemplateList
}

func (o *BootstrapGetBootstrapAgentTemplatesOK) Error() string {
	return fmt.Sprintf("[GET /v2/sentry/bootstrap/template][%d] bootstrapGetBootstrapAgentTemplatesOK  %+v", 200, o.Payload)
}
func (o *BootstrapGetBootstrapAgentTemplatesOK) GetPayload() *models.SentryBootstrapAgentTemplateList {
	return o.Payload
}

func (o *BootstrapGetBootstrapAgentTemplatesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.SentryBootstrapAgentTemplateList)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewBootstrapGetBootstrapAgentTemplatesForbidden creates a BootstrapGetBootstrapAgentTemplatesForbidden with default headers values
func NewBootstrapGetBootstrapAgentTemplatesForbidden() *BootstrapGetBootstrapAgentTemplatesForbidden {
	return &BootstrapGetBootstrapAgentTemplatesForbidden{}
}

/*
	BootstrapGetBootstrapAgentTemplatesForbidden describes a response with status code 403, with default header values.

Returned when the user does not have permission to access the resource.
*/
type BootstrapGetBootstrapAgentTemplatesForbidden struct {
	Payload interface{}
}

func (o *BootstrapGetBootstrapAgentTemplatesForbidden) Error() string {
	return fmt.Sprintf("[GET /v2/sentry/bootstrap/template][%d] bootstrapGetBootstrapAgentTemplatesForbidden  %+v", 403, o.Payload)
}
func (o *BootstrapGetBootstrapAgentTemplatesForbidden) GetPayload() interface{} {
	return o.Payload
}

func (o *BootstrapGetBootstrapAgentTemplatesForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewBootstrapGetBootstrapAgentTemplatesNotFound creates a BootstrapGetBootstrapAgentTemplatesNotFound with default headers values
func NewBootstrapGetBootstrapAgentTemplatesNotFound() *BootstrapGetBootstrapAgentTemplatesNotFound {
	return &BootstrapGetBootstrapAgentTemplatesNotFound{}
}

/*
	BootstrapGetBootstrapAgentTemplatesNotFound describes a response with status code 404, with default header values.

Returned when the resource does not exist.
*/
type BootstrapGetBootstrapAgentTemplatesNotFound struct {
	Payload interface{}
}

func (o *BootstrapGetBootstrapAgentTemplatesNotFound) Error() string {
	return fmt.Sprintf("[GET /v2/sentry/bootstrap/template][%d] bootstrapGetBootstrapAgentTemplatesNotFound  %+v", 404, o.Payload)
}
func (o *BootstrapGetBootstrapAgentTemplatesNotFound) GetPayload() interface{} {
	return o.Payload
}

func (o *BootstrapGetBootstrapAgentTemplatesNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewBootstrapGetBootstrapAgentTemplatesInternalServerError creates a BootstrapGetBootstrapAgentTemplatesInternalServerError with default headers values
func NewBootstrapGetBootstrapAgentTemplatesInternalServerError() *BootstrapGetBootstrapAgentTemplatesInternalServerError {
	return &BootstrapGetBootstrapAgentTemplatesInternalServerError{}
}

/*
	BootstrapGetBootstrapAgentTemplatesInternalServerError describes a response with status code 500, with default header values.

Returned for internal server error
*/
type BootstrapGetBootstrapAgentTemplatesInternalServerError struct {
	Payload interface{}
}

func (o *BootstrapGetBootstrapAgentTemplatesInternalServerError) Error() string {
	return fmt.Sprintf("[GET /v2/sentry/bootstrap/template][%d] bootstrapGetBootstrapAgentTemplatesInternalServerError  %+v", 500, o.Payload)
}
func (o *BootstrapGetBootstrapAgentTemplatesInternalServerError) GetPayload() interface{} {
	return o.Payload
}

func (o *BootstrapGetBootstrapAgentTemplatesInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewBootstrapGetBootstrapAgentTemplatesDefault creates a BootstrapGetBootstrapAgentTemplatesDefault with default headers values
func NewBootstrapGetBootstrapAgentTemplatesDefault(code int) *BootstrapGetBootstrapAgentTemplatesDefault {
	return &BootstrapGetBootstrapAgentTemplatesDefault{
		_statusCode: code,
	}
}

/*
	BootstrapGetBootstrapAgentTemplatesDefault describes a response with status code -1, with default header values.

An unexpected error response.
*/
type BootstrapGetBootstrapAgentTemplatesDefault struct {
	_statusCode int

	Payload *models.GooglerpcStatus
}

// Code gets the status code for the bootstrap get bootstrap agent templates default response
func (o *BootstrapGetBootstrapAgentTemplatesDefault) Code() int {
	return o._statusCode
}

func (o *BootstrapGetBootstrapAgentTemplatesDefault) Error() string {
	return fmt.Sprintf("[GET /v2/sentry/bootstrap/template][%d] Bootstrap_GetBootstrapAgentTemplates default  %+v", o._statusCode, o.Payload)
}
func (o *BootstrapGetBootstrapAgentTemplatesDefault) GetPayload() *models.GooglerpcStatus {
	return o.Payload
}

func (o *BootstrapGetBootstrapAgentTemplatesDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GooglerpcStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
