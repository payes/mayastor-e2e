// Code generated by go-swagger; DO NOT EDIT.

package test_director

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// DeleteTestPlanByIDHandlerFunc turns a function with the right signature into a delete test plan by Id handler
type DeleteTestPlanByIDHandlerFunc func(DeleteTestPlanByIDParams) middleware.Responder

// Handle executing the request and returning a response
func (fn DeleteTestPlanByIDHandlerFunc) Handle(params DeleteTestPlanByIDParams) middleware.Responder {
	return fn(params)
}

// DeleteTestPlanByIDHandler interface for that can handle valid delete test plan by Id params
type DeleteTestPlanByIDHandler interface {
	Handle(DeleteTestPlanByIDParams) middleware.Responder
}

// NewDeleteTestPlanByID creates a new http.Handler for the delete test plan by Id operation
func NewDeleteTestPlanByID(ctx *middleware.Context, handler DeleteTestPlanByIDHandler) *DeleteTestPlanByID {
	return &DeleteTestPlanByID{Context: ctx, Handler: handler}
}

/* DeleteTestPlanByID swagger:route DELETE /td/testplans/{id} test-director deleteTestPlanById

searches for a specific Test Plan by its id

*/
type DeleteTestPlanByID struct {
	Context *middleware.Context
	Handler DeleteTestPlanByIDHandler
}

func (o *DeleteTestPlanByID) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewDeleteTestPlanByIDParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
