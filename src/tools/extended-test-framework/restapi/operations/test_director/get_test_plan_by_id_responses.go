// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"mayastor-e2e/tools/extended-test-framework/models"
)

// GetTestPlanByIDOKCode is the HTTP code returned for type GetTestPlanByIDOK
const GetTestPlanByIDOKCode int = 200

/*GetTestPlanByIDOK Test Plan item returned

swagger:response getTestPlanByIdOK
*/
type GetTestPlanByIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.TestPlan `json:"body,omitempty"`
}

// NewGetTestPlanByIDOK creates GetTestPlanByIDOK with default headers values
func NewGetTestPlanByIDOK() *GetTestPlanByIDOK {

	return &GetTestPlanByIDOK{}
}

// WithPayload adds the payload to the get test plan by Id o k response
func (o *GetTestPlanByIDOK) WithPayload(payload *models.TestPlan) *GetTestPlanByIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get test plan by Id o k response
func (o *GetTestPlanByIDOK) SetPayload(payload *models.TestPlan) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetTestPlanByIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetTestPlanByIDNotFoundCode is the HTTP code returned for type GetTestPlanByIDNotFound
const GetTestPlanByIDNotFoundCode int = 404

/*GetTestPlanByIDNotFound no matching Test Plan found

swagger:response getTestPlanByIdNotFound
*/
type GetTestPlanByIDNotFound struct {
}

// NewGetTestPlanByIDNotFound creates GetTestPlanByIDNotFound with default headers values
func NewGetTestPlanByIDNotFound() *GetTestPlanByIDNotFound {

	return &GetTestPlanByIDNotFound{}
}

// WriteResponse to the client
func (o *GetTestPlanByIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}
