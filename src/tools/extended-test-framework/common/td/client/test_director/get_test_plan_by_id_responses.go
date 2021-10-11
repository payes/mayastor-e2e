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

// GetTestPlanByIDReader is a Reader for the GetTestPlanByID structure.
type GetTestPlanByIDReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetTestPlanByIDReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetTestPlanByIDOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetTestPlanByIDNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetTestPlanByIDOK creates a GetTestPlanByIDOK with default headers values
func NewGetTestPlanByIDOK() *GetTestPlanByIDOK {
	return &GetTestPlanByIDOK{}
}

/* GetTestPlanByIDOK describes a response with status code 200, with default header values.

Test Plan item returned
*/
type GetTestPlanByIDOK struct {
	Payload *models.TestPlan
}

func (o *GetTestPlanByIDOK) Error() string {
	return fmt.Sprintf("[GET /td/testplans/{id}][%d] getTestPlanByIdOK  %+v", 200, o.Payload)
}
func (o *GetTestPlanByIDOK) GetPayload() *models.TestPlan {
	return o.Payload
}

func (o *GetTestPlanByIDOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.TestPlan)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetTestPlanByIDNotFound creates a GetTestPlanByIDNotFound with default headers values
func NewGetTestPlanByIDNotFound() *GetTestPlanByIDNotFound {
	return &GetTestPlanByIDNotFound{}
}

/* GetTestPlanByIDNotFound describes a response with status code 404, with default header values.

no matching Test Plan found
*/
type GetTestPlanByIDNotFound struct {
}

func (o *GetTestPlanByIDNotFound) Error() string {
	return fmt.Sprintf("[GET /td/testplans/{id}][%d] getTestPlanByIdNotFound ", 404)
}

func (o *GetTestPlanByIDNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}