// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"mayastor-e2e/tools/extended-test-framework/common/td/models"
)

// GetTestRunByIDReader is a Reader for the GetTestRunByID structure.
type GetTestRunByIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetTestRunByIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetTestRunByIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetTestRunByIDNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetTestRunByIDOK creates a GetTestRunByIDOK with default headers values
func NewGetTestRunByIDOK() *GetTestRunByIDOK {
	return &GetTestRunByIDOK{}
}

/* GetTestRunByIDOK describes a response with status code 200, with default header values.

A Test Run was returned to the caller
*/
type GetTestRunByIDOK struct {
	Payload *models.TestRun
}

func (o *GetTestRunByIDOK) Error() string {
	return fmt.Sprintf("[GET /td/testruns/{id}][%d] getTestRunByIdOK  %+v", 200, o.Payload)
}
func (o *GetTestRunByIDOK) GetPayload() *models.TestRun {
	return o.Payload
}

func (o *GetTestRunByIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.TestRun)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetTestRunByIDNotFound creates a GetTestRunByIDNotFound with default headers values
func NewGetTestRunByIDNotFound() *GetTestRunByIDNotFound {
	return &GetTestRunByIDNotFound{}
}

/* GetTestRunByIDNotFound describes a response with status code 404, with default header values.

Test Run not found
*/
type GetTestRunByIDNotFound struct {
}

func (o *GetTestRunByIDNotFound) Error() string {
	return fmt.Sprintf("[GET /td/testruns/{id}][%d] getTestRunByIdNotFound ", 404)
}

func (o *GetTestRunByIDNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
