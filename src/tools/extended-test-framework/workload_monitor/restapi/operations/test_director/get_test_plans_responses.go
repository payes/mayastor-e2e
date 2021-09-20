// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
)

// GetTestPlansOKCode is the HTTP code returned for type GetTestPlansOK
const GetTestPlansOKCode int = 200

/*GetTestPlansOK search results available and returned

swagger:response getTestPlansOK
*/
type GetTestPlansOK struct {

	/*
	  In: Body
	*/
	Payload []*models.TestPlan `json:"body,omitempty"`
}

// NewGetTestPlansOK creates GetTestPlansOK with default headers values
func NewGetTestPlansOK() *GetTestPlansOK {

	return &GetTestPlansOK{}
}

// WithPayload adds the payload to the get test plans o k response
func (o *GetTestPlansOK) WithPayload(payload []*models.TestPlan) *GetTestPlansOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get test plans o k response
func (o *GetTestPlansOK) SetPayload(payload []*models.TestPlan) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetTestPlansOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.TestPlan, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GetTestPlansNotFoundCode is the HTTP code returned for type GetTestPlansNotFound
const GetTestPlansNotFoundCode int = 404

/*GetTestPlansNotFound no matching Test Plan(s) found

swagger:response getTestPlansNotFound
*/
type GetTestPlansNotFound struct {
}

// NewGetTestPlansNotFound creates GetTestPlansNotFound with default headers values
func NewGetTestPlansNotFound() *GetTestPlansNotFound {

	return &GetTestPlansNotFound{}
}

// WriteResponse to the client
func (o *GetTestPlansNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}
