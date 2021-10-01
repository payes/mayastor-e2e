// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// TestRunStatusEnum test run status enum
//
// swagger:model TestRunStatusEnum
type TestRunStatusEnum string

func NewTestRunStatusEnum(value TestRunStatusEnum) *TestRunStatusEnum {
	v := value
	return &v
}

const (

	// TestRunStatusEnumNOTSTARTED captures enum value "NOT_STARTED"
	TestRunStatusEnumNOTSTARTED TestRunStatusEnum = "NOT_STARTED"

	// TestRunStatusEnumRUNNING captures enum value "RUNNING"
	TestRunStatusEnumRUNNING TestRunStatusEnum = "RUNNING"

	// TestRunStatusEnumCOMPLETEPASS captures enum value "COMPLETE_PASS"
	TestRunStatusEnumCOMPLETEPASS TestRunStatusEnum = "COMPLETE_PASS"

	// TestRunStatusEnumCOMPLETEFAIL captures enum value "COMPLETE_FAIL"
	TestRunStatusEnumCOMPLETEFAIL TestRunStatusEnum = "COMPLETE_FAIL"
)

// for schema
var testRunStatusEnumEnum []interface{}

func init() {
	var res []TestRunStatusEnum
	if err := json.Unmarshal([]byte(`["NOT_STARTED","RUNNING","COMPLETE_PASS","COMPLETE_FAIL"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		testRunStatusEnumEnum = append(testRunStatusEnumEnum, v)
	}
}

func (m TestRunStatusEnum) validateTestRunStatusEnumEnum(path, location string, value TestRunStatusEnum) error {
	if err := validate.EnumCase(path, location, value, testRunStatusEnumEnum, true); err != nil {
		return err
	}
	return nil
}

// Validate validates this test run status enum
func (m TestRunStatusEnum) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateTestRunStatusEnumEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this test run status enum based on context it is used
func (m TestRunStatusEnum) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}