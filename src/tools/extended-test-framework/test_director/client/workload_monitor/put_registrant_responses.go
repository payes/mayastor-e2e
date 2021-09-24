// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"mayastor-e2e/tools/extended-test-framework/test_director/models"
)

// PutRegistrantReader is a Reader for the PutRegistrant structure.
type PutRegistrantReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PutRegistrantReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPutRegistrantOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewPutRegistrantBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewPutRegistrantOK creates a PutRegistrantOK with default headers values
func NewPutRegistrantOK() *PutRegistrantOK {
	return &PutRegistrantOK{}
}

/* PutRegistrantOK describes a response with status code 200, with default header values.

registrant was registered or updated
*/
type PutRegistrantOK struct {
	Payload *models.Registrant
}

func (o *PutRegistrantOK) Error() string {
	return fmt.Sprintf("[PUT /wm/registrants][%d] putRegistrantOK  %+v", 200, o.Payload)
}
func (o *PutRegistrantOK) GetPayload() *models.Registrant {
	return o.Payload
}

func (o *PutRegistrantOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Registrant)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPutRegistrantBadRequest creates a PutRegistrantBadRequest with default headers values
func NewPutRegistrantBadRequest() *PutRegistrantBadRequest {
	return &PutRegistrantBadRequest{}
}

/* PutRegistrantBadRequest describes a response with status code 400, with default header values.

Bad request (malformed/invalid body content)
*/
type PutRegistrantBadRequest struct {
	Payload *models.RequestOutcome
}

func (o *PutRegistrantBadRequest) Error() string {
	return fmt.Sprintf("[PUT /wm/registrants][%d] putRegistrantBadRequest  %+v", 400, o.Payload)
}
func (o *PutRegistrantBadRequest) GetPayload() *models.RequestOutcome {
	return o.Payload
}

func (o *PutRegistrantBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RequestOutcome)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}