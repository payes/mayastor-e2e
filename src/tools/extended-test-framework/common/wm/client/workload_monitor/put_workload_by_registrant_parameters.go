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

	"mayastor-e2e/tools/extended-test-framework/common/wm/models"
)

// NewPutWorkloadByRegistrantParams creates a new PutWorkloadByRegistrantParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewPutWorkloadByRegistrantParams() *PutWorkloadByRegistrantParams {
	return &PutWorkloadByRegistrantParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewPutWorkloadByRegistrantParamsWithTimeout creates a new PutWorkloadByRegistrantParams object
// with the ability to set a timeout on a request.
func NewPutWorkloadByRegistrantParamsWithTimeout(timeout time.Duration) *PutWorkloadByRegistrantParams {
	return &PutWorkloadByRegistrantParams{
		timeout: timeout,
	}
}

// NewPutWorkloadByRegistrantParamsWithContext creates a new PutWorkloadByRegistrantParams object
// with the ability to set a context for a request.
func NewPutWorkloadByRegistrantParamsWithContext(ctx context.Context) *PutWorkloadByRegistrantParams {
	return &PutWorkloadByRegistrantParams{
		Context: ctx,
	}
}

// NewPutWorkloadByRegistrantParamsWithHTTPClient creates a new PutWorkloadByRegistrantParams object
// with the ability to set a custom HTTPClient for a request.
func NewPutWorkloadByRegistrantParamsWithHTTPClient(client *http.Client) *PutWorkloadByRegistrantParams {
	return &PutWorkloadByRegistrantParams{
		HTTPClient: client,
	}
}

/* PutWorkloadByRegistrantParams contains all the parameters to send to the API endpoint
   for the put workload by registrant operation.

   Typically these are written to a http.Request.
*/
type PutWorkloadByRegistrantParams struct {

	// Body.
	Body *models.WorkloadSpec

	/* Rid.

	   registrant uid

	   Format: uuid
	*/
	Rid strfmt.UUID

	/* Wid.

	   workload uid

	   Format: uuid
	*/
	Wid strfmt.UUID

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the put workload by registrant params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *PutWorkloadByRegistrantParams) WithDefaults() *PutWorkloadByRegistrantParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the put workload by registrant params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *PutWorkloadByRegistrantParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) WithTimeout(timeout time.Duration) *PutWorkloadByRegistrantParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) WithContext(ctx context.Context) *PutWorkloadByRegistrantParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) WithHTTPClient(client *http.Client) *PutWorkloadByRegistrantParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) WithBody(body *models.WorkloadSpec) *PutWorkloadByRegistrantParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) SetBody(body *models.WorkloadSpec) {
	o.Body = body
}

// WithRid adds the rid to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) WithRid(rid strfmt.UUID) *PutWorkloadByRegistrantParams {
	o.SetRid(rid)
	return o
}

// SetRid adds the rid to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) SetRid(rid strfmt.UUID) {
	o.Rid = rid
}

// WithWid adds the wid to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) WithWid(wid strfmt.UUID) *PutWorkloadByRegistrantParams {
	o.SetWid(wid)
	return o
}

// SetWid adds the wid to the put workload by registrant params
func (o *PutWorkloadByRegistrantParams) SetWid(wid strfmt.UUID) {
	o.Wid = wid
}

// WriteToRequest writes these params to a swagger request
func (o *PutWorkloadByRegistrantParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	// path param rid
	if err := r.SetPathParam("rid", o.Rid.String()); err != nil {
		return err
	}

	// path param wid
	if err := r.SetPathParam("wid", o.Wid.String()); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
