// Code generated by go-swagger; DO NOT EDIT.

package bootstrap

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewBootstrapGetBootstrapInfraParams creates a new BootstrapGetBootstrapInfraParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewBootstrapGetBootstrapInfraParams() *BootstrapGetBootstrapInfraParams {
	return &BootstrapGetBootstrapInfraParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewBootstrapGetBootstrapInfraParamsWithTimeout creates a new BootstrapGetBootstrapInfraParams object
// with the ability to set a timeout on a request.
func NewBootstrapGetBootstrapInfraParamsWithTimeout(timeout time.Duration) *BootstrapGetBootstrapInfraParams {
	return &BootstrapGetBootstrapInfraParams{
		timeout: timeout,
	}
}

// NewBootstrapGetBootstrapInfraParamsWithContext creates a new BootstrapGetBootstrapInfraParams object
// with the ability to set a context for a request.
func NewBootstrapGetBootstrapInfraParamsWithContext(ctx context.Context) *BootstrapGetBootstrapInfraParams {
	return &BootstrapGetBootstrapInfraParams{
		Context: ctx,
	}
}

// NewBootstrapGetBootstrapInfraParamsWithHTTPClient creates a new BootstrapGetBootstrapInfraParams object
// with the ability to set a custom HTTPClient for a request.
func NewBootstrapGetBootstrapInfraParamsWithHTTPClient(client *http.Client) *BootstrapGetBootstrapInfraParams {
	return &BootstrapGetBootstrapInfraParams{
		HTTPClient: client,
	}
}

/*
BootstrapGetBootstrapInfraParams contains all the parameters to send to the API endpoint

	for the bootstrap get bootstrap infra operation.

	Typically these are written to a http.Request.
*/
type BootstrapGetBootstrapInfraParams struct {

	/* APIVersion.

	   API Version. API Version of the resource

	   Default: "infra.k8smgmt.io/v3"
	*/
	APIVersion *string

	/* Kind.

	   Kind. Kind of the resource

	   Default: "BootstrapInfra"
	*/
	Kind *string

	/* MetadataDescription.

	   Description. description of the resource
	*/
	MetadataDescription *string

	/* MetadataDisplayName.

	   Display Name. display name of the resource
	*/
	MetadataDisplayName *string

	// MetadataID.
	MetadataID *string

	// MetadataModifiedAt.
	//
	// Format: date-time
	MetadataModifiedAt *strfmt.DateTime

	/* MetadataName.

	   name of the resource
	*/
	MetadataName string

	/* MetadataOrganization.

	   Organization. Organization to which the resource belongs
	*/
	MetadataOrganization *string

	/* MetadataPartner.

	   Partner. Partner to which the resource belongs
	*/
	MetadataPartner *string

	/* MetadataProject.

	   Project. Project of the resource
	*/
	MetadataProject *string

	// SpecCaCert.
	SpecCaCert *string

	// SpecCaKey.
	SpecCaKey *string

	// SpecCaKeyPass.
	SpecCaKeyPass *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the bootstrap get bootstrap infra params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *BootstrapGetBootstrapInfraParams) WithDefaults() *BootstrapGetBootstrapInfraParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the bootstrap get bootstrap infra params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *BootstrapGetBootstrapInfraParams) SetDefaults() {
	var (
		aPIVersionDefault = string("infra.k8smgmt.io/v3")

		kindDefault = string("BootstrapInfra")
	)

	val := BootstrapGetBootstrapInfraParams{
		APIVersion: &aPIVersionDefault,
		Kind:       &kindDefault,
	}

	val.timeout = o.timeout
	val.Context = o.Context
	val.HTTPClient = o.HTTPClient
	*o = val
}

// WithTimeout adds the timeout to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithTimeout(timeout time.Duration) *BootstrapGetBootstrapInfraParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithContext(ctx context.Context) *BootstrapGetBootstrapInfraParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithHTTPClient(client *http.Client) *BootstrapGetBootstrapInfraParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAPIVersion adds the aPIVersion to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithAPIVersion(aPIVersion *string) *BootstrapGetBootstrapInfraParams {
	o.SetAPIVersion(aPIVersion)
	return o
}

// SetAPIVersion adds the apiVersion to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetAPIVersion(aPIVersion *string) {
	o.APIVersion = aPIVersion
}

// WithKind adds the kind to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithKind(kind *string) *BootstrapGetBootstrapInfraParams {
	o.SetKind(kind)
	return o
}

// SetKind adds the kind to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetKind(kind *string) {
	o.Kind = kind
}

// WithMetadataDescription adds the metadataDescription to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataDescription(metadataDescription *string) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataDescription(metadataDescription)
	return o
}

// SetMetadataDescription adds the metadataDescription to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataDescription(metadataDescription *string) {
	o.MetadataDescription = metadataDescription
}

// WithMetadataDisplayName adds the metadataDisplayName to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataDisplayName(metadataDisplayName *string) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataDisplayName(metadataDisplayName)
	return o
}

// SetMetadataDisplayName adds the metadataDisplayName to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataDisplayName(metadataDisplayName *string) {
	o.MetadataDisplayName = metadataDisplayName
}

// WithMetadataID adds the metadataID to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataID(metadataID *string) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataID(metadataID)
	return o
}

// SetMetadataID adds the metadataId to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataID(metadataID *string) {
	o.MetadataID = metadataID
}

// WithMetadataModifiedAt adds the metadataModifiedAt to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataModifiedAt(metadataModifiedAt *strfmt.DateTime) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataModifiedAt(metadataModifiedAt)
	return o
}

// SetMetadataModifiedAt adds the metadataModifiedAt to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataModifiedAt(metadataModifiedAt *strfmt.DateTime) {
	o.MetadataModifiedAt = metadataModifiedAt
}

// WithMetadataName adds the metadataName to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataName(metadataName string) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataName(metadataName)
	return o
}

// SetMetadataName adds the metadataName to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataName(metadataName string) {
	o.MetadataName = metadataName
}

// WithMetadataOrganization adds the metadataOrganization to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataOrganization(metadataOrganization *string) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataOrganization(metadataOrganization)
	return o
}

// SetMetadataOrganization adds the metadataOrganization to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataOrganization(metadataOrganization *string) {
	o.MetadataOrganization = metadataOrganization
}

// WithMetadataPartner adds the metadataPartner to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataPartner(metadataPartner *string) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataPartner(metadataPartner)
	return o
}

// SetMetadataPartner adds the metadataPartner to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataPartner(metadataPartner *string) {
	o.MetadataPartner = metadataPartner
}

// WithMetadataProject adds the metadataProject to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithMetadataProject(metadataProject *string) *BootstrapGetBootstrapInfraParams {
	o.SetMetadataProject(metadataProject)
	return o
}

// SetMetadataProject adds the metadataProject to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetMetadataProject(metadataProject *string) {
	o.MetadataProject = metadataProject
}

// WithSpecCaCert adds the specCaCert to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithSpecCaCert(specCaCert *string) *BootstrapGetBootstrapInfraParams {
	o.SetSpecCaCert(specCaCert)
	return o
}

// SetSpecCaCert adds the specCaCert to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetSpecCaCert(specCaCert *string) {
	o.SpecCaCert = specCaCert
}

// WithSpecCaKey adds the specCaKey to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithSpecCaKey(specCaKey *string) *BootstrapGetBootstrapInfraParams {
	o.SetSpecCaKey(specCaKey)
	return o
}

// SetSpecCaKey adds the specCaKey to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetSpecCaKey(specCaKey *string) {
	o.SpecCaKey = specCaKey
}

// WithSpecCaKeyPass adds the specCaKeyPass to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) WithSpecCaKeyPass(specCaKeyPass *string) *BootstrapGetBootstrapInfraParams {
	o.SetSpecCaKeyPass(specCaKeyPass)
	return o
}

// SetSpecCaKeyPass adds the specCaKeyPass to the bootstrap get bootstrap infra params
func (o *BootstrapGetBootstrapInfraParams) SetSpecCaKeyPass(specCaKeyPass *string) {
	o.SpecCaKeyPass = specCaKeyPass
}

// WriteToRequest writes these params to a swagger request
func (o *BootstrapGetBootstrapInfraParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.APIVersion != nil {

		// query param apiVersion
		var qrAPIVersion string

		if o.APIVersion != nil {
			qrAPIVersion = *o.APIVersion
		}
		qAPIVersion := qrAPIVersion
		if qAPIVersion != "" {

			if err := r.SetQueryParam("apiVersion", qAPIVersion); err != nil {
				return err
			}
		}
	}

	if o.Kind != nil {

		// query param kind
		var qrKind string

		if o.Kind != nil {
			qrKind = *o.Kind
		}
		qKind := qrKind
		if qKind != "" {

			if err := r.SetQueryParam("kind", qKind); err != nil {
				return err
			}
		}
	}

	if o.MetadataDescription != nil {

		// query param metadata.description
		var qrMetadataDescription string

		if o.MetadataDescription != nil {
			qrMetadataDescription = *o.MetadataDescription
		}
		qMetadataDescription := qrMetadataDescription
		if qMetadataDescription != "" {

			if err := r.SetQueryParam("metadata.description", qMetadataDescription); err != nil {
				return err
			}
		}
	}

	if o.MetadataDisplayName != nil {

		// query param metadata.displayName
		var qrMetadataDisplayName string

		if o.MetadataDisplayName != nil {
			qrMetadataDisplayName = *o.MetadataDisplayName
		}
		qMetadataDisplayName := qrMetadataDisplayName
		if qMetadataDisplayName != "" {

			if err := r.SetQueryParam("metadata.displayName", qMetadataDisplayName); err != nil {
				return err
			}
		}
	}

	if o.MetadataID != nil {

		// query param metadata.id
		var qrMetadataID string

		if o.MetadataID != nil {
			qrMetadataID = *o.MetadataID
		}
		qMetadataID := qrMetadataID
		if qMetadataID != "" {

			if err := r.SetQueryParam("metadata.id", qMetadataID); err != nil {
				return err
			}
		}
	}

	if o.MetadataModifiedAt != nil {

		// query param metadata.modifiedAt
		var qrMetadataModifiedAt strfmt.DateTime

		if o.MetadataModifiedAt != nil {
			qrMetadataModifiedAt = *o.MetadataModifiedAt
		}
		qMetadataModifiedAt := qrMetadataModifiedAt.String()
		if qMetadataModifiedAt != "" {

			if err := r.SetQueryParam("metadata.modifiedAt", qMetadataModifiedAt); err != nil {
				return err
			}
		}
	}

	// path param metadata.name
	if err := r.SetPathParam("metadata.name", o.MetadataName); err != nil {
		return err
	}

	if o.MetadataOrganization != nil {

		// query param metadata.organization
		var qrMetadataOrganization string

		if o.MetadataOrganization != nil {
			qrMetadataOrganization = *o.MetadataOrganization
		}
		qMetadataOrganization := qrMetadataOrganization
		if qMetadataOrganization != "" {

			if err := r.SetQueryParam("metadata.organization", qMetadataOrganization); err != nil {
				return err
			}
		}
	}

	if o.MetadataPartner != nil {

		// query param metadata.partner
		var qrMetadataPartner string

		if o.MetadataPartner != nil {
			qrMetadataPartner = *o.MetadataPartner
		}
		qMetadataPartner := qrMetadataPartner
		if qMetadataPartner != "" {

			if err := r.SetQueryParam("metadata.partner", qMetadataPartner); err != nil {
				return err
			}
		}
	}

	if o.MetadataProject != nil {

		// query param metadata.project
		var qrMetadataProject string

		if o.MetadataProject != nil {
			qrMetadataProject = *o.MetadataProject
		}
		qMetadataProject := qrMetadataProject
		if qMetadataProject != "" {

			if err := r.SetQueryParam("metadata.project", qMetadataProject); err != nil {
				return err
			}
		}
	}

	if o.SpecCaCert != nil {

		// query param spec.caCert
		var qrSpecCaCert string

		if o.SpecCaCert != nil {
			qrSpecCaCert = *o.SpecCaCert
		}
		qSpecCaCert := qrSpecCaCert
		if qSpecCaCert != "" {

			if err := r.SetQueryParam("spec.caCert", qSpecCaCert); err != nil {
				return err
			}
		}
	}

	if o.SpecCaKey != nil {

		// query param spec.caKey
		var qrSpecCaKey string

		if o.SpecCaKey != nil {
			qrSpecCaKey = *o.SpecCaKey
		}
		qSpecCaKey := qrSpecCaKey
		if qSpecCaKey != "" {

			if err := r.SetQueryParam("spec.caKey", qSpecCaKey); err != nil {
				return err
			}
		}
	}

	if o.SpecCaKeyPass != nil {

		// query param spec.caKeyPass
		var qrSpecCaKeyPass string

		if o.SpecCaKeyPass != nil {
			qrSpecCaKeyPass = *o.SpecCaKeyPass
		}
		qSpecCaKeyPass := qrSpecCaKeyPass
		if qSpecCaKeyPass != "" {

			if err := r.SetQueryParam("spec.caKeyPass", qSpecCaKeyPass); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
