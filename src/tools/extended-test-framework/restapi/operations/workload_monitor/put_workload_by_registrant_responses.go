// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"mayastor-e2e/tools/extended-test-framework/models"
)

// PutWorkloadByRegistrantOKCode is the HTTP code returned for type PutWorkloadByRegistrantOK
const PutWorkloadByRegistrantOKCode int = 200

/*PutWorkloadByRegistrantOK Workload registered/updated

swagger:response putWorkloadByRegistrantOK
*/
type PutWorkloadByRegistrantOK struct {

	/*
	  In: Body
	*/
	Payload *models.Workload `json:"body,omitempty"`
}

// NewPutWorkloadByRegistrantOK creates PutWorkloadByRegistrantOK with default headers values
func NewPutWorkloadByRegistrantOK() *PutWorkloadByRegistrantOK {

	return &PutWorkloadByRegistrantOK{}
}

// WithPayload adds the payload to the put workload by registrant o k response
func (o *PutWorkloadByRegistrantOK) WithPayload(payload *models.Workload) *PutWorkloadByRegistrantOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put workload by registrant o k response
func (o *PutWorkloadByRegistrantOK) SetPayload(payload *models.Workload) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutWorkloadByRegistrantOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// PutWorkloadByRegistrantBadRequestCode is the HTTP code returned for type PutWorkloadByRegistrantBadRequest
const PutWorkloadByRegistrantBadRequestCode int = 400

/*PutWorkloadByRegistrantBadRequest Bad request (malformed/invalid resource id)

swagger:response putWorkloadByRegistrantBadRequest
*/
type PutWorkloadByRegistrantBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.RequestOutcome `json:"body,omitempty"`
}

// NewPutWorkloadByRegistrantBadRequest creates PutWorkloadByRegistrantBadRequest with default headers values
func NewPutWorkloadByRegistrantBadRequest() *PutWorkloadByRegistrantBadRequest {

	return &PutWorkloadByRegistrantBadRequest{}
}

// WithPayload adds the payload to the put workload by registrant bad request response
func (o *PutWorkloadByRegistrantBadRequest) WithPayload(payload *models.RequestOutcome) *PutWorkloadByRegistrantBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the put workload by registrant bad request response
func (o *PutWorkloadByRegistrantBadRequest) SetPayload(payload *models.RequestOutcome) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PutWorkloadByRegistrantBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
