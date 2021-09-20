// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
)

// DeleteTestPlansOKCode is the HTTP code returned for type DeleteTestPlansOK
const DeleteTestPlansOKCode int = 200

/*DeleteTestPlansOK Returns deleted Test Plan count, which may be zero

swagger:response deleteTestPlansOK
*/
type DeleteTestPlansOK struct {

	/*
	  In: Body
	*/
	Payload *models.RequestOutcome `json:"body,omitempty"`
}

// NewDeleteTestPlansOK creates DeleteTestPlansOK with default headers values
func NewDeleteTestPlansOK() *DeleteTestPlansOK {

	return &DeleteTestPlansOK{}
}

// WithPayload adds the payload to the delete test plans o k response
func (o *DeleteTestPlansOK) WithPayload(payload *models.RequestOutcome) *DeleteTestPlansOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete test plans o k response
func (o *DeleteTestPlansOK) SetPayload(payload *models.RequestOutcome) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteTestPlansOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
