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

// DeleteTestRunByIDReader is a Reader for the DeleteTestRunByID structure.
type DeleteTestRunByIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteTestRunByIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewDeleteTestRunByIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewDeleteTestRunByIDNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDeleteTestRunByIDOK creates a DeleteTestRunByIDOK with default headers values
func NewDeleteTestRunByIDOK() *DeleteTestRunByIDOK {
	return &DeleteTestRunByIDOK{}
}

/* DeleteTestRunByIDOK describes a response with status code 200, with default header values.

Returns the deleted Test Run
*/
type DeleteTestRunByIDOK struct {
	Payload []*models.TestRun
}

func (o *DeleteTestRunByIDOK) Error() string {
	return fmt.Sprintf("[DELETE /td/testruns/{id}][%d] deleteTestRunByIdOK  %+v", 200, o.Payload)
}
func (o *DeleteTestRunByIDOK) GetPayload() []*models.TestRun {
	return o.Payload
}

func (o *DeleteTestRunByIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteTestRunByIDNotFound creates a DeleteTestRunByIDNotFound with default headers values
func NewDeleteTestRunByIDNotFound() *DeleteTestRunByIDNotFound {
	return &DeleteTestRunByIDNotFound{}
}

/* DeleteTestRunByIDNotFound describes a response with status code 404, with default header values.

Test Run not found
*/
type DeleteTestRunByIDNotFound struct {
}

func (o *DeleteTestRunByIDNotFound) Error() string {
	return fmt.Sprintf("[DELETE /td/testruns/{id}][%d] deleteTestRunByIdNotFound ", 404)
}

func (o *DeleteTestRunByIDNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
