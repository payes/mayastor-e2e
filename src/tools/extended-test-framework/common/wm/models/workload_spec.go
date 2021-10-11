// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// WorkloadSpec workload spec
//
// swagger:model WorkloadSpec
type WorkloadSpec struct {

	// violations
	// Required: true
	Violations []WorkloadViolationEnum `json:"violations"`
}

// Validate validates this workload spec
func (m *WorkloadSpec) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateViolations(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *WorkloadSpec) validateViolations(formats strfmt.Registry) error {

	if err := validate.Required("violations", "body", m.Violations); err != nil {
		return err
	}

	for i := 0; i < len(m.Violations); i++ {

		if err := m.Violations[i].Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("violations" + "." + strconv.Itoa(i))
			}
			return err
		}

	}

	return nil
}

// ContextValidate validate this workload spec based on the context it is used
func (m *WorkloadSpec) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateViolations(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *WorkloadSpec) contextValidateViolations(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Violations); i++ {

		if err := m.Violations[i].ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("violations" + "." + strconv.Itoa(i))
			}
			return err
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *WorkloadSpec) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *WorkloadSpec) UnmarshalBinary(b []byte) error {
	var res WorkloadSpec
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}