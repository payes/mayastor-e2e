// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"mayastor-e2e/tools/extended-test-framework/test_director/models"
)

// PutTestPlanByIDOKCode is the HTTP code returned for type PutTestPlanByIDOK
const PutTestPlanByIDOKCode int = 200

/*PutTestPlanByIDOK Test Plan was registered or updated

swagger:response putTestPlanByIdOK
*/
type PutTestPlanByIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.TestPlan `json:"body,omitempty"`
}

// NewPutTestPlanByIDOK creates PutTestPlanByIDOK with default headers values
func NewPutTestPlanByIDOK() *PutTestPlanByIDOK {

	return &PutTestPlanByIDOK{}
}

// WithPayload adds the payload to the put test plan by Id o k response
func (o *PutTestPlanByIDOK) WithPayload(payload *models.TestPlan) *PutTestPlanByIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put test plan by Id o k response
func (o *PutTestPlanByIDOK) SetPayload(payload *models.TestPlan) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutTestPlanByIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutTestPlanByIDBadRequestCode is the HTTP code returned for type PutTestPlanByIDBadRequest
const PutTestPlanByIDBadRequestCode int = 400

/*PutTestPlanByIDBadRequest Bad request (malformed/invalid body content)

swagger:response putTestPlanByIdBadRequest
*/
type PutTestPlanByIDBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.RequestOutcome `json:"body,omitempty"`
}

// NewPutTestPlanByIDBadRequest creates PutTestPlanByIDBadRequest with default headers values
func NewPutTestPlanByIDBadRequest() *PutTestPlanByIDBadRequest {

	return &PutTestPlanByIDBadRequest{}
}

// WithPayload adds the payload to the put test plan by Id bad request response
func (o *PutTestPlanByIDBadRequest) WithPayload(payload *models.RequestOutcome) *PutTestPlanByIDBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put test plan by Id bad request response
func (o *PutTestPlanByIDBadRequest) SetPayload(payload *models.RequestOutcome) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutTestPlanByIDBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}