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

// TestPlanStatusEnum test plan status enum
//
// swagger:model TestPlanStatusEnum
type TestPlanStatusEnum string

func NewTestPlanStatusEnum(value TestPlanStatusEnum) *TestPlanStatusEnum {
	v := value
	return &v
}

const (

	// TestPlanStatusEnumNOTSTARTED captures enum value "NOT_STARTED"
	TestPlanStatusEnumNOTSTARTED TestPlanStatusEnum = "NOT_STARTED"

	// TestPlanStatusEnumRUNNING captures enum value "RUNNING"
	TestPlanStatusEnumRUNNING TestPlanStatusEnum = "RUNNING"

	// TestPlanStatusEnumCOMPLETEPASS captures enum value "COMPLETE_PASS"
	TestPlanStatusEnumCOMPLETEPASS TestPlanStatusEnum = "COMPLETE_PASS"

	// TestPlanStatusEnumCOMPLETEFAIL captures enum value "COMPLETE_FAIL"
	TestPlanStatusEnumCOMPLETEFAIL TestPlanStatusEnum = "COMPLETE_FAIL"
)

// for schema
var testPlanStatusEnumEnum []interface{}

func init() {
	var res []TestPlanStatusEnum
	if err := json.Unmarshal([]byte(`["NOT_STARTED","RUNNING","COMPLETE_PASS","COMPLETE_FAIL"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		testPlanStatusEnumEnum = append(testPlanStatusEnumEnum, v)
	}
}

func (m TestPlanStatusEnum) validateTestPlanStatusEnumEnum(path, location string, value TestPlanStatusEnum) error {
	if err := validate.EnumCase(path, location, value, testPlanStatusEnumEnum, true); err != nil {
		return err
	}
	return nil
}

// Validate validates this test plan status enum
func (m TestPlanStatusEnum) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateTestPlanStatusEnumEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this test plan status enum based on context it is used
func (m TestPlanStatusEnum) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
