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

// GetWorkloadsByRegistrantReader is a Reader for the GetWorkloadsByRegistrant structure.
type GetWorkloadsByRegistrantReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetWorkloadsByRegistrantReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetWorkloadsByRegistrantOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetWorkloadsByRegistrantBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetWorkloadsByRegistrantNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetWorkloadsByRegistrantOK creates a GetWorkloadsByRegistrantOK with default headers values
func NewGetWorkloadsByRegistrantOK() *GetWorkloadsByRegistrantOK {
	return &GetWorkloadsByRegistrantOK{}
}

/* GetWorkloadsByRegistrantOK describes a response with status code 200, with default header values.

corresponding Workload item(s) returned to caller
*/
type GetWorkloadsByRegistrantOK struct {
	Payload []*models.Workload
}

func (o *GetWorkloadsByRegistrantOK) Error() string {
	return fmt.Sprintf("[GET /wm/registrants/{rid}][%d] getWorkloadsByRegistrantOK  %+v", 200, o.Payload)
}
func (o *GetWorkloadsByRegistrantOK) GetPayload() []*models.Workload {
	return o.Payload
}

func (o *GetWorkloadsByRegistrantOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetWorkloadsByRegistrantBadRequest creates a GetWorkloadsByRegistrantBadRequest with default headers values
func NewGetWorkloadsByRegistrantBadRequest() *GetWorkloadsByRegistrantBadRequest {
	return &GetWorkloadsByRegistrantBadRequest{}
}

/* GetWorkloadsByRegistrantBadRequest describes a response with status code 400, with default header values.

Bad request (malformed/invalid resource id)
*/
type GetWorkloadsByRegistrantBadRequest struct {
	Payload *models.RequestOutcome
}

func (o *GetWorkloadsByRegistrantBadRequest) Error() string {
	return fmt.Sprintf("[GET /wm/registrants/{rid}][%d] getWorkloadsByRegistrantBadRequest  %+v", 400, o.Payload)
}
func (o *GetWorkloadsByRegistrantBadRequest) GetPayload() *models.RequestOutcome {
	return o.Payload
}

func (o *GetWorkloadsByRegistrantBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.RequestOutcome)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetWorkloadsByRegistrantNotFound creates a GetWorkloadsByRegistrantNotFound with default headers values
func NewGetWorkloadsByRegistrantNotFound() *GetWorkloadsByRegistrantNotFound {
	return &GetWorkloadsByRegistrantNotFound{}
}

/* GetWorkloadsByRegistrantNotFound describes a response with status code 404, with default header values.

no corresponding workload(s) found
*/
type GetWorkloadsByRegistrantNotFound struct {
}

func (o *GetWorkloadsByRegistrantNotFound) Error() string {
	return fmt.Sprintf("[GET /wm/registrants/{rid}][%d] getWorkloadsByRegistrantNotFound ", 404)
}

func (o *GetWorkloadsByRegistrantNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
