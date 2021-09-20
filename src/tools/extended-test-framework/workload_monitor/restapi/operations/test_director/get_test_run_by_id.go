// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetTestRunByIDHandlerFunc turns a function with the right signature into a get test run by Id handler
type GetTestRunByIDHandlerFunc func(GetTestRunByIDParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetTestRunByIDHandlerFunc) Handle(params GetTestRunByIDParams) middleware.Responder {
	return fn(params)
}

// GetTestRunByIDHandler interface for that can handle valid get test run by Id params
type GetTestRunByIDHandler interface {
	Handle(GetTestRunByIDParams) middleware.Responder
}

// NewGetTestRunByID creates a new http.Handler for the get test run by Id operation
func NewGetTestRunByID(ctx *middleware.Context, handler GetTestRunByIDHandler) *GetTestRunByID {
	return &GetTestRunByID{Context: ctx, Handler: handler}
}

/* GetTestRunByID swagger:route GET /td/testruns/{id} test-director getTestRunById

returns a Test Run with the corresponding id

*/
type GetTestRunByID struct {
	Context *middleware.Context
	Handler GetTestRunByIDHandler
}

func (o *GetTestRunByID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetTestRunByIDParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
