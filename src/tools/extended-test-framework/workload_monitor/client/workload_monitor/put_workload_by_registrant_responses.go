// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
)

// PutWorkloadByRegistrantReader is a Reader for the PutWorkloadByRegistrant structure.
type PutWorkloadByRegistrantReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PutWorkloadByRegistrantReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPutWorkloadByRegistrantOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewPutWorkloadByRegistrantBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewPutWorkloadByRegistrantOK creates a PutWorkloadByRegistrantOK with default headers values
func NewPutWorkloadByRegistrantOK() *PutWorkloadByRegistrantOK {
	return &PutWorkloadByRegistrantOK{}
}

/* PutWorkloadByRegistrantOK describes a response with status code 200, with default header values.

Workload registered/updated
*/
type PutWorkloadByRegistrantOK struct {
	Payload *models.Workload
}

func (o *PutWorkloadByRegistrantOK) Error() string {
	return fmt.Sprintf("[PUT /wm/registrants/{rid}/workloads/{wid}][%d] putWorkloadByRegistrantOK  %+v", 200, o.Payload)
}
func (o *PutWorkloadByRegistrantOK) GetPayload() *models.Workload {
	return o.Payload
}

func (o *PutWorkloadByRegistrantOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Workload)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPutWorkloadByRegistrantBadRequest creates a PutWorkloadByRegistrantBadRequest with default headers values
func NewPutWorkloadByRegistrantBadRequest() *PutWorkloadByRegistrantBadRequest {
	return &PutWorkloadByRegistrantBadRequest{}
}

/* PutWorkloadByRegistrantBadRequest describes a response with status code 400, with default header values.

Bad request (malformed/invalid resource id)
*/
type PutWorkloadByRegistrantBadRequest struct {
	Payload *models.RequestOutcome
}

func (o *PutWorkloadByRegistrantBadRequest) Error() string {
	return fmt.Sprintf("[PUT /wm/registrants/{rid}/workloads/{wid}][%d] putWorkloadByRegistrantBadRequest  %+v", 400, o.Payload)
}
func (o *PutWorkloadByRegistrantBadRequest) GetPayload() *models.RequestOutcome {
	return o.Payload
}

func (o *PutWorkloadByRegistrantBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RequestOutcome)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
