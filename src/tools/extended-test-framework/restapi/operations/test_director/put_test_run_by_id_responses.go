// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"mayastor-e2e/tools/extended-test-framework/models"
)

// PutTestRunByIDOKCode is the HTTP code returned for type PutTestRunByIDOK
const PutTestRunByIDOKCode int = 200

/*PutTestRunByIDOK Test Run Registered/Updated

swagger:response putTestRunByIdOK
*/
type PutTestRunByIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.TestRun `json:"body,omitempty"`
}

// NewPutTestRunByIDOK creates PutTestRunByIDOK with default headers values
func NewPutTestRunByIDOK() *PutTestRunByIDOK {

	return &PutTestRunByIDOK{}
}

// WithPayload adds the payload to the put test run by Id o k response
func (o *PutTestRunByIDOK) WithPayload(payload *models.TestRun) *PutTestRunByIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put test run by Id o k response
func (o *PutTestRunByIDOK) SetPayload(payload *models.TestRun) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutTestRunByIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutTestRunByIDBadRequestCode is the HTTP code returned for type PutTestRunByIDBadRequest
const PutTestRunByIDBadRequestCode int = 400

/*PutTestRunByIDBadRequest Bad request

swagger:response putTestRunByIdBadRequest
*/
type PutTestRunByIDBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.RequestOutcome `json:"body,omitempty"`
}

// NewPutTestRunByIDBadRequest creates PutTestRunByIDBadRequest with default headers values
func NewPutTestRunByIDBadRequest() *PutTestRunByIDBadRequest {

	return &PutTestRunByIDBadRequest{}
}

// WithPayload adds the payload to the put test run by Id bad request response
func (o *PutTestRunByIDBadRequest) WithPayload(payload *models.RequestOutcome) *PutTestRunByIDBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put test run by Id bad request response
func (o *PutTestRunByIDBadRequest) SetPayload(payload *models.RequestOutcome) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutTestRunByIDBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutTestRunByIDForbiddenCode is the HTTP code returned for type PutTestRunByIDForbidden
const PutTestRunByIDForbiddenCode int = 403

/*PutTestRunByIDForbidden The request was refused

swagger:response putTestRunByIdForbidden
*/
type PutTestRunByIDForbidden struct {

	/*
	  In: Body
	*/
	Payload *models.RequestOutcome `json:"body,omitempty"`
}

// NewPutTestRunByIDForbidden creates PutTestRunByIDForbidden with default headers values
func NewPutTestRunByIDForbidden() *PutTestRunByIDForbidden {

	return &PutTestRunByIDForbidden{}
}

// WithPayload adds the payload to the put test run by Id forbidden response
func (o *PutTestRunByIDForbidden) WithPayload(payload *models.RequestOutcome) *PutTestRunByIDForbidden {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put test run by Id forbidden response
func (o *PutTestRunByIDForbidden) SetPayload(payload *models.RequestOutcome) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutTestRunByIDForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(403)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
