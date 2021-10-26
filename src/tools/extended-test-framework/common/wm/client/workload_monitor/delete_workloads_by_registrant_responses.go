// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"mayastor-e2e/tools/extended-test-framework/common/wm/models"
)

// DeleteWorkloadsByRegistrantReader is a Reader for the DeleteWorkloadsByRegistrant structure.
type DeleteWorkloadsByRegistrantReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteWorkloadsByRegistrantReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewDeleteWorkloadsByRegistrantOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewDeleteWorkloadsByRegistrantBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewDeleteWorkloadsByRegistrantNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDeleteWorkloadsByRegistrantOK creates a DeleteWorkloadsByRegistrantOK with default headers values
func NewDeleteWorkloadsByRegistrantOK() *DeleteWorkloadsByRegistrantOK {
	return &DeleteWorkloadsByRegistrantOK{}
}

/* DeleteWorkloadsByRegistrantOK describes a response with status code 200, with default header values.

Workload(s) deleted
*/
type DeleteWorkloadsByRegistrantOK struct {
	Payload *models.RequestOutcome
}

func (o *DeleteWorkloadsByRegistrantOK) Error() string {
	return fmt.Sprintf("[DELETE /wm/registrants/{rid}][%d] deleteWorkloadsByRegistrantOK  %+v", 200, o.Payload)
}
func (o *DeleteWorkloadsByRegistrantOK) GetPayload() *models.RequestOutcome {
	return o.Payload
}

func (o *DeleteWorkloadsByRegistrantOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RequestOutcome)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteWorkloadsByRegistrantBadRequest creates a DeleteWorkloadsByRegistrantBadRequest with default headers values
func NewDeleteWorkloadsByRegistrantBadRequest() *DeleteWorkloadsByRegistrantBadRequest {
	return &DeleteWorkloadsByRegistrantBadRequest{}
}

/* DeleteWorkloadsByRegistrantBadRequest describes a response with status code 400, with default header values.

Bad request (malformed/invalid resource id)
*/
type DeleteWorkloadsByRegistrantBadRequest struct {
	Payload *models.RequestOutcome
}

func (o *DeleteWorkloadsByRegistrantBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /wm/registrants/{rid}][%d] deleteWorkloadsByRegistrantBadRequest  %+v", 400, o.Payload)
}
func (o *DeleteWorkloadsByRegistrantBadRequest) GetPayload() *models.RequestOutcome {
	return o.Payload
}

func (o *DeleteWorkloadsByRegistrantBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RequestOutcome)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteWorkloadsByRegistrantNotFound creates a DeleteWorkloadsByRegistrantNotFound with default headers values
func NewDeleteWorkloadsByRegistrantNotFound() *DeleteWorkloadsByRegistrantNotFound {
	return &DeleteWorkloadsByRegistrantNotFound{}
}

/* DeleteWorkloadsByRegistrantNotFound describes a response with status code 404, with default header values.

resource not found
*/
type DeleteWorkloadsByRegistrantNotFound struct {
}

func (o *DeleteWorkloadsByRegistrantNotFound) Error() string {
	return fmt.Sprintf("[DELETE /wm/registrants/{rid}][%d] deleteWorkloadsByRegistrantNotFound ", 404)
}

func (o *DeleteWorkloadsByRegistrantNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
