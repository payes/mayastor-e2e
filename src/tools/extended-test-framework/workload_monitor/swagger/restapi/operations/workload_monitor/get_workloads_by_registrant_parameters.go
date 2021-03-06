// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// NewGetWorkloadsByRegistrantParams creates a new GetWorkloadsByRegistrantParams object
//
// There are no default values defined in the spec.
func NewGetWorkloadsByRegistrantParams() GetWorkloadsByRegistrantParams {

	return GetWorkloadsByRegistrantParams{}
}

// GetWorkloadsByRegistrantParams contains all the bound params for the get workloads by registrant operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetWorkloadsByRegistrant
type GetWorkloadsByRegistrantParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*registrant uid
	  Required: true
	  In: path
	*/
	Rid strfmt.UUID
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetWorkloadsByRegistrantParams() beforehand.
func (o *GetWorkloadsByRegistrantParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rRid, rhkRid, _ := route.Params.GetOK("rid")
	if err := o.bindRid(rRid, rhkRid, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindRid binds and validates parameter Rid from path.
func (o *GetWorkloadsByRegistrantParams) bindRid(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	// Format: uuid
	value, err := formats.Parse("uuid", raw)
	if err != nil {
		return errors.InvalidType("rid", "path", "strfmt.UUID", raw)
	}
	o.Rid = *(value.(*strfmt.UUID))

	if err := o.validateRid(formats); err != nil {
		return err
	}

	return nil
}

// validateRid carries on validations for parameter Rid
func (o *GetWorkloadsByRegistrantParams) validateRid(formats strfmt.Registry) error {

	if err := validate.FormatOf("rid", "path", "uuid", o.Rid.String(), formats); err != nil {
		return err
	}
	return nil
}
