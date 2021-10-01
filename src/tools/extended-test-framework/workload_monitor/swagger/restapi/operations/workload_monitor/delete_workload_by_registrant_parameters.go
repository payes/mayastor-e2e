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

// NewDeleteWorkloadByRegistrantParams creates a new DeleteWorkloadByRegistrantParams object
//
// There are no default values defined in the spec.
func NewDeleteWorkloadByRegistrantParams() DeleteWorkloadByRegistrantParams {

	return DeleteWorkloadByRegistrantParams{}
}

// DeleteWorkloadByRegistrantParams contains all the bound params for the delete workload by registrant operation
// typically these are obtained from a http.Request
//
// swagger:parameters DeleteWorkloadByRegistrant
type DeleteWorkloadByRegistrantParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*registrant uid
	  Required: true
	  In: path
	*/
	Rid strfmt.UUID
	/*workload uid
	  Required: true
	  In: path
	*/
	Wid strfmt.UUID
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewDeleteWorkloadByRegistrantParams() beforehand.
func (o *DeleteWorkloadByRegistrantParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rRid, rhkRid, _ := route.Params.GetOK("rid")
	if err := o.bindRid(rRid, rhkRid, route.Formats); err != nil {
		res = append(res, err)
	}

	rWid, rhkWid, _ := route.Params.GetOK("wid")
	if err := o.bindWid(rWid, rhkWid, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindRid binds and validates parameter Rid from path.
func (o *DeleteWorkloadByRegistrantParams) bindRid(rawData []string, hasKey bool, formats strfmt.Registry) error {
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
func (o *DeleteWorkloadByRegistrantParams) validateRid(formats strfmt.Registry) error {

	if err := validate.FormatOf("rid", "path", "uuid", o.Rid.String(), formats); err != nil {
		return err
	}
	return nil
}

// bindWid binds and validates parameter Wid from path.
func (o *DeleteWorkloadByRegistrantParams) bindWid(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	// Format: uuid
	value, err := formats.Parse("uuid", raw)
	if err != nil {
		return errors.InvalidType("wid", "path", "strfmt.UUID", raw)
	}
	o.Wid = *(value.(*strfmt.UUID))

	if err := o.validateWid(formats); err != nil {
		return err
	}

	return nil
}

// validateWid carries on validations for parameter Wid
func (o *DeleteWorkloadByRegistrantParams) validateWid(formats strfmt.Registry) error {

	if err := validate.FormatOf("wid", "path", "uuid", o.Wid.String(), formats); err != nil {
		return err
	}
	return nil
}