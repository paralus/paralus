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
	"github.com/go-openapi/swag"
)

// NewBootstrapGetBootstrapAgentTemplatesParams creates a new BootstrapGetBootstrapAgentTemplatesParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewBootstrapGetBootstrapAgentTemplatesParams() *BootstrapGetBootstrapAgentTemplatesParams {
	return &BootstrapGetBootstrapAgentTemplatesParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewBootstrapGetBootstrapAgentTemplatesParamsWithTimeout creates a new BootstrapGetBootstrapAgentTemplatesParams object
// with the ability to set a timeout on a request.
func NewBootstrapGetBootstrapAgentTemplatesParamsWithTimeout(timeout time.Duration) *BootstrapGetBootstrapAgentTemplatesParams {
	return &BootstrapGetBootstrapAgentTemplatesParams{
		timeout: timeout,
	}
}

// NewBootstrapGetBootstrapAgentTemplatesParamsWithContext creates a new BootstrapGetBootstrapAgentTemplatesParams object
// with the ability to set a context for a request.
func NewBootstrapGetBootstrapAgentTemplatesParamsWithContext(ctx context.Context) *BootstrapGetBootstrapAgentTemplatesParams {
	return &BootstrapGetBootstrapAgentTemplatesParams{
		Context: ctx,
	}
}

// NewBootstrapGetBootstrapAgentTemplatesParamsWithHTTPClient creates a new BootstrapGetBootstrapAgentTemplatesParams object
// with the ability to set a custom HTTPClient for a request.
func NewBootstrapGetBootstrapAgentTemplatesParamsWithHTTPClient(client *http.Client) *BootstrapGetBootstrapAgentTemplatesParams {
	return &BootstrapGetBootstrapAgentTemplatesParams{
		HTTPClient: client,
	}
}

/*
BootstrapGetBootstrapAgentTemplatesParams contains all the parameters to send to the API endpoint

	for the bootstrap get bootstrap agent templates operation.

	Typically these are written to a http.Request.
*/
type BootstrapGetBootstrapAgentTemplatesParams struct {

	// ID.
	ID *string

	// Account.
	Account *string

	// BlueprintRef.
	BlueprintRef *string

	// ClusterID.
	ClusterID *string

	// Count.
	//
	// Format: int64
	Count *string

	// Deleted.
	Deleted *bool

	/* DisplayName.

	   displayName only used for update queries to set displayName (READONLY).
	*/
	DisplayName *string

	// Extended.
	Extended *bool

	/* GlobalScope.

	   globalScope sets partnerID,organizationID,projectID = 0.
	*/
	GlobalScope *bool

	// Groups.
	Groups []string

	/* IgnoreScopeDefault.

	     ignoreScopeDefault ignores default values for partnerID, organizationID and
	projectID.
	*/
	IgnoreScopeDefault *bool

	// IsSSOUser.
	IsSSOUser *bool

	// Limit.
	//
	// Format: int64
	Limit *string

	/* Name.

	     name is unique ID of a resource along with (partnerID, organizationID,
	projectID).
	*/
	Name *string

	// Offset.
	//
	// Format: int64
	Offset *string

	// Order.
	Order *string

	// OrderBy.
	OrderBy *string

	// Organization.
	Organization *string

	// Partner.
	Partner *string

	// Project.
	Project *string

	// PublishedVersion.
	PublishedVersion *string

	/* Selector.

	   selector is used to filter the labels of a resource.
	*/
	Selector *string

	/* URLScope.

	   urlScope is supposed to be passed in the URL as kind/HashID(value).
	*/
	URLScope *string

	// Username.
	Username *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the bootstrap get bootstrap agent templates params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithDefaults() *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the bootstrap get bootstrap agent templates params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithTimeout(timeout time.Duration) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithContext(ctx context.Context) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithHTTPClient(client *http.Client) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithID(id *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetID(id *string) {
	o.ID = id
}

// WithAccount adds the account to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithAccount(account *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetAccount(account)
	return o
}

// SetAccount adds the account to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetAccount(account *string) {
	o.Account = account
}

// WithBlueprintRef adds the blueprintRef to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithBlueprintRef(blueprintRef *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetBlueprintRef(blueprintRef)
	return o
}

// SetBlueprintRef adds the blueprintRef to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetBlueprintRef(blueprintRef *string) {
	o.BlueprintRef = blueprintRef
}

// WithClusterID adds the clusterID to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithClusterID(clusterID *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetClusterID(clusterID)
	return o
}

// SetClusterID adds the clusterId to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetClusterID(clusterID *string) {
	o.ClusterID = clusterID
}

// WithCount adds the count to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithCount(count *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetCount(count)
	return o
}

// SetCount adds the count to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetCount(count *string) {
	o.Count = count
}

// WithDeleted adds the deleted to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithDeleted(deleted *bool) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetDeleted(deleted)
	return o
}

// SetDeleted adds the deleted to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetDeleted(deleted *bool) {
	o.Deleted = deleted
}

// WithDisplayName adds the displayName to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithDisplayName(displayName *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetDisplayName(displayName)
	return o
}

// SetDisplayName adds the displayName to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetDisplayName(displayName *string) {
	o.DisplayName = displayName
}

// WithExtended adds the extended to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithExtended(extended *bool) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetExtended(extended)
	return o
}

// SetExtended adds the extended to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetExtended(extended *bool) {
	o.Extended = extended
}

// WithGlobalScope adds the globalScope to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithGlobalScope(globalScope *bool) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetGlobalScope(globalScope)
	return o
}

// SetGlobalScope adds the globalScope to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetGlobalScope(globalScope *bool) {
	o.GlobalScope = globalScope
}

// WithGroups adds the groups to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithGroups(groups []string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetGroups(groups)
	return o
}

// SetGroups adds the groups to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetGroups(groups []string) {
	o.Groups = groups
}

// WithIgnoreScopeDefault adds the ignoreScopeDefault to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithIgnoreScopeDefault(ignoreScopeDefault *bool) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetIgnoreScopeDefault(ignoreScopeDefault)
	return o
}

// SetIgnoreScopeDefault adds the ignoreScopeDefault to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetIgnoreScopeDefault(ignoreScopeDefault *bool) {
	o.IgnoreScopeDefault = ignoreScopeDefault
}

// WithIsSSOUser adds the isSSOUser to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithIsSSOUser(isSSOUser *bool) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetIsSSOUser(isSSOUser)
	return o
}

// SetIsSSOUser adds the isSSOUser to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetIsSSOUser(isSSOUser *bool) {
	o.IsSSOUser = isSSOUser
}

// WithLimit adds the limit to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithLimit(limit *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetLimit(limit)
	return o
}

// SetLimit adds the limit to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetLimit(limit *string) {
	o.Limit = limit
}

// WithName adds the name to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithName(name *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetName(name *string) {
	o.Name = name
}

// WithOffset adds the offset to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithOffset(offset *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetOffset(offset)
	return o
}

// SetOffset adds the offset to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetOffset(offset *string) {
	o.Offset = offset
}

// WithOrder adds the order to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithOrder(order *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetOrder(order)
	return o
}

// SetOrder adds the order to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetOrder(order *string) {
	o.Order = order
}

// WithOrderBy adds the orderBy to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithOrderBy(orderBy *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetOrderBy(orderBy)
	return o
}

// SetOrderBy adds the orderBy to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetOrderBy(orderBy *string) {
	o.OrderBy = orderBy
}

// WithOrganization adds the organization to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithOrganization(organization *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetOrganization(organization)
	return o
}

// SetOrganization adds the organization to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetOrganization(organization *string) {
	o.Organization = organization
}

// WithPartner adds the partner to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithPartner(partner *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetPartner(partner)
	return o
}

// SetPartner adds the partner to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetPartner(partner *string) {
	o.Partner = partner
}

// WithProject adds the project to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithProject(project *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetProject(project)
	return o
}

// SetProject adds the project to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetProject(project *string) {
	o.Project = project
}

// WithPublishedVersion adds the publishedVersion to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithPublishedVersion(publishedVersion *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetPublishedVersion(publishedVersion)
	return o
}

// SetPublishedVersion adds the publishedVersion to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetPublishedVersion(publishedVersion *string) {
	o.PublishedVersion = publishedVersion
}

// WithSelector adds the selector to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithSelector(selector *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetSelector(selector)
	return o
}

// SetSelector adds the selector to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetSelector(selector *string) {
	o.Selector = selector
}

// WithURLScope adds the uRLScope to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithURLScope(uRLScope *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetURLScope(uRLScope)
	return o
}

// SetURLScope adds the urlScope to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetURLScope(uRLScope *string) {
	o.URLScope = uRLScope
}

// WithUsername adds the username to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) WithUsername(username *string) *BootstrapGetBootstrapAgentTemplatesParams {
	o.SetUsername(username)
	return o
}

// SetUsername adds the username to the bootstrap get bootstrap agent templates params
func (o *BootstrapGetBootstrapAgentTemplatesParams) SetUsername(username *string) {
	o.Username = username
}

// WriteToRequest writes these params to a swagger request
func (o *BootstrapGetBootstrapAgentTemplatesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.ID != nil {

		// query param ID
		var qrID string

		if o.ID != nil {
			qrID = *o.ID
		}
		qID := qrID
		if qID != "" {

			if err := r.SetQueryParam("ID", qID); err != nil {
				return err
			}
		}
	}

	if o.Account != nil {

		// query param account
		var qrAccount string

		if o.Account != nil {
			qrAccount = *o.Account
		}
		qAccount := qrAccount
		if qAccount != "" {

			if err := r.SetQueryParam("account", qAccount); err != nil {
				return err
			}
		}
	}

	if o.BlueprintRef != nil {

		// query param blueprintRef
		var qrBlueprintRef string

		if o.BlueprintRef != nil {
			qrBlueprintRef = *o.BlueprintRef
		}
		qBlueprintRef := qrBlueprintRef
		if qBlueprintRef != "" {

			if err := r.SetQueryParam("blueprintRef", qBlueprintRef); err != nil {
				return err
			}
		}
	}

	if o.ClusterID != nil {

		// query param clusterID
		var qrClusterID string

		if o.ClusterID != nil {
			qrClusterID = *o.ClusterID
		}
		qClusterID := qrClusterID
		if qClusterID != "" {

			if err := r.SetQueryParam("clusterID", qClusterID); err != nil {
				return err
			}
		}
	}

	if o.Count != nil {

		// query param count
		var qrCount string

		if o.Count != nil {
			qrCount = *o.Count
		}
		qCount := qrCount
		if qCount != "" {

			if err := r.SetQueryParam("count", qCount); err != nil {
				return err
			}
		}
	}

	if o.Deleted != nil {

		// query param deleted
		var qrDeleted bool

		if o.Deleted != nil {
			qrDeleted = *o.Deleted
		}
		qDeleted := swag.FormatBool(qrDeleted)
		if qDeleted != "" {

			if err := r.SetQueryParam("deleted", qDeleted); err != nil {
				return err
			}
		}
	}

	if o.DisplayName != nil {

		// query param displayName
		var qrDisplayName string

		if o.DisplayName != nil {
			qrDisplayName = *o.DisplayName
		}
		qDisplayName := qrDisplayName
		if qDisplayName != "" {

			if err := r.SetQueryParam("displayName", qDisplayName); err != nil {
				return err
			}
		}
	}

	if o.Extended != nil {

		// query param extended
		var qrExtended bool

		if o.Extended != nil {
			qrExtended = *o.Extended
		}
		qExtended := swag.FormatBool(qrExtended)
		if qExtended != "" {

			if err := r.SetQueryParam("extended", qExtended); err != nil {
				return err
			}
		}
	}

	if o.GlobalScope != nil {

		// query param globalScope
		var qrGlobalScope bool

		if o.GlobalScope != nil {
			qrGlobalScope = *o.GlobalScope
		}
		qGlobalScope := swag.FormatBool(qrGlobalScope)
		if qGlobalScope != "" {

			if err := r.SetQueryParam("globalScope", qGlobalScope); err != nil {
				return err
			}
		}
	}

	if o.Groups != nil {

		// binding items for groups
		joinedGroups := o.bindParamGroups(reg)

		// query array param groups
		if err := r.SetQueryParam("groups", joinedGroups...); err != nil {
			return err
		}
	}

	if o.IgnoreScopeDefault != nil {

		// query param ignoreScopeDefault
		var qrIgnoreScopeDefault bool

		if o.IgnoreScopeDefault != nil {
			qrIgnoreScopeDefault = *o.IgnoreScopeDefault
		}
		qIgnoreScopeDefault := swag.FormatBool(qrIgnoreScopeDefault)
		if qIgnoreScopeDefault != "" {

			if err := r.SetQueryParam("ignoreScopeDefault", qIgnoreScopeDefault); err != nil {
				return err
			}
		}
	}

	if o.IsSSOUser != nil {

		// query param isSSOUser
		var qrIsSSOUser bool

		if o.IsSSOUser != nil {
			qrIsSSOUser = *o.IsSSOUser
		}
		qIsSSOUser := swag.FormatBool(qrIsSSOUser)
		if qIsSSOUser != "" {

			if err := r.SetQueryParam("isSSOUser", qIsSSOUser); err != nil {
				return err
			}
		}
	}

	if o.Limit != nil {

		// query param limit
		var qrLimit string

		if o.Limit != nil {
			qrLimit = *o.Limit
		}
		qLimit := qrLimit
		if qLimit != "" {

			if err := r.SetQueryParam("limit", qLimit); err != nil {
				return err
			}
		}
	}

	if o.Name != nil {

		// query param name
		var qrName string

		if o.Name != nil {
			qrName = *o.Name
		}
		qName := qrName
		if qName != "" {

			if err := r.SetQueryParam("name", qName); err != nil {
				return err
			}
		}
	}

	if o.Offset != nil {

		// query param offset
		var qrOffset string

		if o.Offset != nil {
			qrOffset = *o.Offset
		}
		qOffset := qrOffset
		if qOffset != "" {

			if err := r.SetQueryParam("offset", qOffset); err != nil {
				return err
			}
		}
	}

	if o.Order != nil {

		// query param order
		var qrOrder string

		if o.Order != nil {
			qrOrder = *o.Order
		}
		qOrder := qrOrder
		if qOrder != "" {

			if err := r.SetQueryParam("order", qOrder); err != nil {
				return err
			}
		}
	}

	if o.OrderBy != nil {

		// query param orderBy
		var qrOrderBy string

		if o.OrderBy != nil {
			qrOrderBy = *o.OrderBy
		}
		qOrderBy := qrOrderBy
		if qOrderBy != "" {

			if err := r.SetQueryParam("orderBy", qOrderBy); err != nil {
				return err
			}
		}
	}

	if o.Organization != nil {

		// query param organization
		var qrOrganization string

		if o.Organization != nil {
			qrOrganization = *o.Organization
		}
		qOrganization := qrOrganization
		if qOrganization != "" {

			if err := r.SetQueryParam("organization", qOrganization); err != nil {
				return err
			}
		}
	}

	if o.Partner != nil {

		// query param partner
		var qrPartner string

		if o.Partner != nil {
			qrPartner = *o.Partner
		}
		qPartner := qrPartner
		if qPartner != "" {

			if err := r.SetQueryParam("partner", qPartner); err != nil {
				return err
			}
		}
	}

	if o.Project != nil {

		// query param project
		var qrProject string

		if o.Project != nil {
			qrProject = *o.Project
		}
		qProject := qrProject
		if qProject != "" {

			if err := r.SetQueryParam("project", qProject); err != nil {
				return err
			}
		}
	}

	if o.PublishedVersion != nil {

		// query param publishedVersion
		var qrPublishedVersion string

		if o.PublishedVersion != nil {
			qrPublishedVersion = *o.PublishedVersion
		}
		qPublishedVersion := qrPublishedVersion
		if qPublishedVersion != "" {

			if err := r.SetQueryParam("publishedVersion", qPublishedVersion); err != nil {
				return err
			}
		}
	}

	if o.Selector != nil {

		// query param selector
		var qrSelector string

		if o.Selector != nil {
			qrSelector = *o.Selector
		}
		qSelector := qrSelector
		if qSelector != "" {

			if err := r.SetQueryParam("selector", qSelector); err != nil {
				return err
			}
		}
	}

	if o.URLScope != nil {

		// query param urlScope
		var qrURLScope string

		if o.URLScope != nil {
			qrURLScope = *o.URLScope
		}
		qURLScope := qrURLScope
		if qURLScope != "" {

			if err := r.SetQueryParam("urlScope", qURLScope); err != nil {
				return err
			}
		}
	}

	if o.Username != nil {

		// query param username
		var qrUsername string

		if o.Username != nil {
			qrUsername = *o.Username
		}
		qUsername := qrUsername
		if qUsername != "" {

			if err := r.SetQueryParam("username", qUsername); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindParamBootstrapGetBootstrapAgentTemplates binds the parameter groups
func (o *BootstrapGetBootstrapAgentTemplatesParams) bindParamGroups(formats strfmt.Registry) []string {
	groupsIR := o.Groups

	var groupsIC []string
	for _, groupsIIR := range groupsIR { // explode []string

		groupsIIV := groupsIIR // string as string
		groupsIC = append(groupsIC, groupsIIV)
	}

	// items.CollectionFormat: "multi"
	groupsIS := swag.JoinByFormat(groupsIC, "multi")

	return groupsIS
}
