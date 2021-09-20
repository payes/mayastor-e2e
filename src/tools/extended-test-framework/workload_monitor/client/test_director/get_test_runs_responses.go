// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
)

// GetTestRunsReader is a Reader for the GetTestRuns structure.
type GetTestRunsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetTestRunsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetTestRunsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetTestRunsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetTestRunsOK creates a GetTestRunsOK with default headers values
func NewGetTestRunsOK() *GetTestRunsOK {
	return &GetTestRunsOK{}
}

/* GetTestRunsOK describes a response with status code 200, with default header values.

Test object(s) returned
*/
type GetTestRunsOK struct {
	Payload []*models.TestRun
}

func (o *GetTestRunsOK) Error() string {
	return fmt.Sprintf("[GET /td/testRuns][%d] getTestRunsOK  %+v", 200, o.Payload)
}
func (o *GetTestRunsOK) GetPayload() []*models.TestRun {
	return o.Payload
}

func (o *GetTestRunsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetTestRunsNotFound creates a GetTestRunsNotFound with default headers values
func NewGetTestRunsNotFound() *GetTestRunsNotFound {
	return &GetTestRunsNotFound{}
}

/* GetTestRunsNotFound describes a response with status code 404, with default header values.

no matching Test(s) found
*/
type GetTestRunsNotFound struct {
}

func (o *GetTestRunsNotFound) Error() string {
	return fmt.Sprintf("[GET /td/testRuns][%d] getTestRunsNotFound ", 404)
}

func (o *GetTestRunsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
