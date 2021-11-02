// Code generated by go-swagger; DO NOT EDIT.

package test_director

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

// NewDeleteTestPlansParams creates a new DeleteTestPlansParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewDeleteTestPlansParams() *DeleteTestPlansParams {
	return &DeleteTestPlansParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewDeleteTestPlansParamsWithTimeout creates a new DeleteTestPlansParams object
// with the ability to set a timeout on a request.
func NewDeleteTestPlansParamsWithTimeout(timeout time.Duration) *DeleteTestPlansParams {
	return &DeleteTestPlansParams{
		timeout: timeout,
	}
}

// NewDeleteTestPlansParamsWithContext creates a new DeleteTestPlansParams object
// with the ability to set a context for a request.
func NewDeleteTestPlansParamsWithContext(ctx context.Context) *DeleteTestPlansParams {
	return &DeleteTestPlansParams{
		Context: ctx,
	}
}

// NewDeleteTestPlansParamsWithHTTPClient creates a new DeleteTestPlansParams object
// with the ability to set a custom HTTPClient for a request.
func NewDeleteTestPlansParamsWithHTTPClient(client *http.Client) *DeleteTestPlansParams {
	return &DeleteTestPlansParams{
		HTTPClient: client,
	}
}

/* DeleteTestPlansParams contains all the parameters to send to the API endpoint
   for the delete test plans operation.

   Typically these are written to a http.Request.
*/
type DeleteTestPlansParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the delete test plans params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteTestPlansParams) WithDefaults() *DeleteTestPlansParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the delete test plans params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DeleteTestPlansParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the delete test plans params
func (o *DeleteTestPlansParams) WithTimeout(timeout time.Duration) *DeleteTestPlansParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the delete test plans params
func (o *DeleteTestPlansParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the delete test plans params
func (o *DeleteTestPlansParams) WithContext(ctx context.Context) *DeleteTestPlansParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the delete test plans params
func (o *DeleteTestPlansParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the delete test plans params
func (o *DeleteTestPlansParams) WithHTTPClient(client *http.Client) *DeleteTestPlansParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the delete test plans params
func (o *DeleteTestPlansParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *DeleteTestPlansParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
