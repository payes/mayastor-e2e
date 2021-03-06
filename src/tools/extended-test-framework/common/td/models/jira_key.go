// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// JiraKey jira key
// Example: MQ-123
//
// swagger:model JiraKey
type JiraKey string

// Validate validates this jira key
func (m JiraKey) Validate(formats strfmt.Registry) error {
	var res []error

	if err := validate.Pattern("", "body", string(m), `^[A-Z]{2,3}-\d{1,4}$`); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this jira key based on context it is used
func (m JiraKey) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
