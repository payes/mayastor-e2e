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

// NewGetWorkloadsByRegistrantParams creates a new GetWorkloadsByRegistrantParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetWorkloadsByRegistrantParams() *GetWorkloadsByRegistrantParams {
	return &GetWorkloadsByRegistrantParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetWorkloadsByRegistrantParamsWithTimeout creates a new GetWorkloadsByRegistrantParams object
// with the ability to set a timeout on a request.
func NewGetWorkloadsByRegistrantParamsWithTimeout(timeout time.Duration) *GetWorkloadsByRegistrantParams {
	return &GetWorkloadsByRegistrantParams{
		timeout: timeout,
	}
}

// NewGetWorkloadsByRegistrantParamsWithContext creates a new GetWorkloadsByRegistrantParams object
// with the ability to set a context for a request.
func NewGetWorkloadsByRegistrantParamsWithContext(ctx context.Context) *GetWorkloadsByRegistrantParams {
	return &GetWorkloadsByRegistrantParams{
		Context: ctx,
	}
}

// NewGetWorkloadsByRegistrantParamsWithHTTPClient creates a new GetWorkloadsByRegistrantParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetWorkloadsByRegistrantParamsWithHTTPClient(client *http.Client) *GetWorkloadsByRegistrantParams {
	return &GetWorkloadsByRegistrantParams{
		HTTPClient: client,
	}
}

/* GetWorkloadsByRegistrantParams contains all the parameters to send to the API endpoint
   for the get workloads by registrant operation.

   Typically these are written to a http.Request.
*/
type GetWorkloadsByRegistrantParams struct {

	/* Rid.

	   registrant uid

	   Format: uuid
	*/
	Rid strfmt.UUID

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get workloads by registrant params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetWorkloadsByRegistrantParams) WithDefaults() *GetWorkloadsByRegistrantParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get workloads by registrant params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetWorkloadsByRegistrantParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) WithTimeout(timeout time.Duration) *GetWorkloadsByRegistrantParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) WithContext(ctx context.Context) *GetWorkloadsByRegistrantParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) WithHTTPClient(client *http.Client) *GetWorkloadsByRegistrantParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithRid adds the rid to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) WithRid(rid strfmt.UUID) *GetWorkloadsByRegistrantParams {
	o.SetRid(rid)
	return o
}

// SetRid adds the rid to the get workloads by registrant params
func (o *GetWorkloadsByRegistrantParams) SetRid(rid strfmt.UUID) {
	o.Rid = rid
}

// WriteToRequest writes these params to a swagger request
func (o *GetWorkloadsByRegistrantParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param rid
	if err := r.SetPathParam("rid", o.Rid.String()); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
