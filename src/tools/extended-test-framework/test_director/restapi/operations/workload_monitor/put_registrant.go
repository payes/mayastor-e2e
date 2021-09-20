// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// PutRegistrantHandlerFunc turns a function with the right signature into a put registrant handler
type PutRegistrantHandlerFunc func(PutRegistrantParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PutRegistrantHandlerFunc) Handle(params PutRegistrantParams) middleware.Responder {
	return fn(params)
}

// PutRegistrantHandler interface for that can handle valid put registrant params
type PutRegistrantHandler interface {
	Handle(PutRegistrantParams) middleware.Responder
}

// NewPutRegistrant creates a new http.Handler for the put registrant operation
func NewPutRegistrant(ctx *middleware.Context, handler PutRegistrantHandler) *PutRegistrant {
	return &PutRegistrant{Context: ctx, Handler: handler}
}

/* PutRegistrant swagger:route PUT /wm/registrants workload-monitor putRegistrant

registers or updates a registrant with the workload-monitor

*/
type PutRegistrant struct {
	Context *middleware.Context
	Handler PutRegistrantHandler
}

func (o *PutRegistrant) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewPutRegistrantParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
