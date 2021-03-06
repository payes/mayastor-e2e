// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"test-director/models"
)

// GetTestRunByIDOKCode is the HTTP code returned for type GetTestRunByIDOK
const GetTestRunByIDOKCode int = 200

/*GetTestRunByIDOK A Test Run was returned to the caller

swagger:response getTestRunByIdOK
*/
type GetTestRunByIDOK struct {

	/*
	  In: Body
	*/
	Payload *models.TestRun `json:"body,omitempty"`
}

// NewGetTestRunByIDOK creates GetTestRunByIDOK with default headers values
func NewGetTestRunByIDOK() *GetTestRunByIDOK {

	return &GetTestRunByIDOK{}
}

// WithPayload adds the payload to the get test run by Id o k response
func (o *GetTestRunByIDOK) WithPayload(payload *models.TestRun) *GetTestRunByIDOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get test run by Id o k response
func (o *GetTestRunByIDOK) SetPayload(payload *models.TestRun) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetTestRunByIDOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetTestRunByIDNotFoundCode is the HTTP code returned for type GetTestRunByIDNotFound
const GetTestRunByIDNotFoundCode int = 404

/*GetTestRunByIDNotFound Test Run not found

swagger:response getTestRunByIdNotFound
*/
type GetTestRunByIDNotFound struct {
}

// NewGetTestRunByIDNotFound creates GetTestRunByIDNotFound with default headers values
func NewGetTestRunByIDNotFound() *GetTestRunByIDNotFound {

	return &GetTestRunByIDNotFound{}
}

// WriteResponse to the client
func (o *GetTestRunByIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(404)
}
