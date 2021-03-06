// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// DeleteTestRunByIDHandlerFunc turns a function with the right signature into a delete test run by Id handler
type DeleteTestRunByIDHandlerFunc func(DeleteTestRunByIDParams) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteTestRunByIDHandlerFunc) Handle(params DeleteTestRunByIDParams) middleware.Responder {
	return fn(params)
}

// DeleteTestRunByIDHandler interface for that can handle valid delete test run by Id params
type DeleteTestRunByIDHandler interface {
	Handle(DeleteTestRunByIDParams) middleware.Responder
}

// NewDeleteTestRunByID creates a new http.Handler for the delete test run by Id operation
func NewDeleteTestRunByID(ctx *middleware.Context, handler DeleteTestRunByIDHandler) *DeleteTestRunByID {
	return &DeleteTestRunByID{Context: ctx, Handler: handler}
}

/* DeleteTestRunByID swagger:route DELETE /td/testruns/{id} test-director deleteTestRunById

returns a Test Run with the corresponding id

*/
type DeleteTestRunByID struct {
	Context *middleware.Context
	Handler DeleteTestRunByIDHandler
}

func (o *DeleteTestRunByID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewDeleteTestRunByIDParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
