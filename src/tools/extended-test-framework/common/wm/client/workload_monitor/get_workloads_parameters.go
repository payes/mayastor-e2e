// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

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

// NewGetWorkloadsParams creates a new GetWorkloadsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetWorkloadsParams() *GetWorkloadsParams {
	return &GetWorkloadsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetWorkloadsParamsWithTimeout creates a new GetWorkloadsParams object
// with the ability to set a timeout on a request.
func NewGetWorkloadsParamsWithTimeout(timeout time.Duration) *GetWorkloadsParams {
	return &GetWorkloadsParams{
		timeout: timeout,
	}
}

// NewGetWorkloadsParamsWithContext creates a new GetWorkloadsParams object
// with the ability to set a context for a request.
func NewGetWorkloadsParamsWithContext(ctx context.Context) *GetWorkloadsParams {
	return &GetWorkloadsParams{
		Context: ctx,
	}
}

// NewGetWorkloadsParamsWithHTTPClient creates a new GetWorkloadsParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetWorkloadsParamsWithHTTPClient(client *http.Client) *GetWorkloadsParams {
	return &GetWorkloadsParams{
		HTTPClient: client,
	}
}

/* GetWorkloadsParams contains all the parameters to send to the API endpoint
   for the get workloads operation.

   Typically these are written to a http.Request.
*/
type GetWorkloadsParams struct {

	/* Name.

	   workload (pod) name
	*/
	Name *string

	/* Namespace.

	   workload (pod) namespace
	*/
	Namespace *string

	/* RegistrantID.

	   metadata.uid of Pod which registered the workload

	   Format: uuid
	*/
	RegistrantID *strfmt.UUID

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get workloads params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetWorkloadsParams) WithDefaults() *GetWorkloadsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get workloads params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetWorkloadsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get workloads params
func (o *GetWorkloadsParams) WithTimeout(timeout time.Duration) *GetWorkloadsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get workloads params
func (o *GetWorkloadsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get workloads params
func (o *GetWorkloadsParams) WithContext(ctx context.Context) *GetWorkloadsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get workloads params
func (o *GetWorkloadsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get workloads params
func (o *GetWorkloadsParams) WithHTTPClient(client *http.Client) *GetWorkloadsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get workloads params
func (o *GetWorkloadsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the get workloads params
func (o *GetWorkloadsParams) WithName(name *string) *GetWorkloadsParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the get workloads params
func (o *GetWorkloadsParams) SetName(name *string) {
	o.Name = name
}

// WithNamespace adds the namespace to the get workloads params
func (o *GetWorkloadsParams) WithNamespace(namespace *string) *GetWorkloadsParams {
	o.SetNamespace(namespace)
	return o
}

// SetNamespace adds the namespace to the get workloads params
func (o *GetWorkloadsParams) SetNamespace(namespace *string) {
	o.Namespace = namespace
}

// WithRegistrantID adds the registrantID to the get workloads params
func (o *GetWorkloadsParams) WithRegistrantID(registrantID *strfmt.UUID) *GetWorkloadsParams {
	o.SetRegistrantID(registrantID)
	return o
}

// SetRegistrantID adds the registrantId to the get workloads params
func (o *GetWorkloadsParams) SetRegistrantID(registrantID *strfmt.UUID) {
	o.RegistrantID = registrantID
}

// WriteToRequest writes these params to a swagger request
func (o *GetWorkloadsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

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

	if o.Namespace != nil {

		// query param namespace
		var qrNamespace string

		if o.Namespace != nil {
			qrNamespace = *o.Namespace
		}
		qNamespace := qrNamespace
		if qNamespace != "" {

			if err := r.SetQueryParam("namespace", qNamespace); err != nil {
				return err
			}
		}
	}

	if o.RegistrantID != nil {

		// query param registrantId
		var qrRegistrantID strfmt.UUID

		if o.RegistrantID != nil {
			qrRegistrantID = *o.RegistrantID
		}
		qRegistrantID := qrRegistrantID.String()
		if qRegistrantID != "" {

			if err := r.SetQueryParam("registrantId", qRegistrantID); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
