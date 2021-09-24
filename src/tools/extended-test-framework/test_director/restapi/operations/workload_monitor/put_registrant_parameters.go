// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"io"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/validate"

	"mayastor-e2e/tools/extended-test-framework/test_director/models"
)

// NewPutRegistrantParams creates a new PutRegistrantParams object
//
// There are no default values defined in the spec.
func NewPutRegistrantParams() PutRegistrantParams {

	return PutRegistrantParams{}
}

// PutRegistrantParams contains all the bound params for the put registrant operation
// typically these are obtained from a http.Request
//
// swagger:parameters PutRegistrant
type PutRegistrantParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*registrant uid
	  Required: true
	  In: body
	*/
	Rid *models.Registrant
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewPutRegistrantParams() beforehand.
func (o *PutRegistrantParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.Registrant
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			if err == io.EOF {
				res = append(res, errors.Required("rid", "body", ""))
			} else {
				res = append(res, errors.NewParseError("rid", "body", "", err))
			}
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			ctx := validate.WithOperationRequest(context.Background())
			if err := body.ContextValidate(ctx, route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Rid = &body
			}
		}
	} else {
		res = append(res, errors.Required("rid", "body", ""))
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}