// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
)

// GetTestRunsOKCode is the HTTP code returned for type GetTestRunsOK
const GetTestRunsOKCode int = 200

/*GetTestRunsOK Test object(s) returned

swagger:response getTestRunsOK
*/
type GetTestRunsOK struct {

	/*
	  In: Body
	*/
	Payload []*models.TestRun `json:"body,omitempty"`
}

// NewGetTestRunsOK creates GetTestRunsOK with default headers values
func NewGetTestRunsOK() *GetTestRunsOK {

	return &GetTestRunsOK{}
}

// WithPayload adds the payload to the get test runs o k response
func (o *GetTestRunsOK) WithPayload(payload []*models.TestRun) *GetTestRunsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get test runs o k response
func (o *GetTestRunsOK) SetPayload(payload []*models.TestRun) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetTestRunsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.TestRun, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GetTestRunsNotFoundCode is the HTTP code returned for type GetTestRunsNotFound
const GetTestRunsNotFoundCode int = 404

/*GetTestRunsNotFound no matching Test(s) found

swagger:response getTestRunsNotFound
*/
type GetTestRunsNotFound struct {
}

// NewGetTestRunsNotFound creates GetTestRunsNotFound with default headers values
func NewGetTestRunsNotFound() *GetTestRunsNotFound {

	return &GetTestRunsNotFound{}
}

// WriteResponse to the client
func (o *GetTestRunsNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}
