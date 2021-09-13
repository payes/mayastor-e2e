// Code generated by go-swagger; DO NOT EDIT.

package workload_monitor

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetWorkloadsHandlerFunc turns a function with the right signature into a get workloads handler
type GetWorkloadsHandlerFunc func(GetWorkloadsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetWorkloadsHandlerFunc) Handle(params GetWorkloadsParams) middleware.Responder {
	return fn(params)
}

// GetWorkloadsHandler interface for that can handle valid get workloads params
type GetWorkloadsHandler interface {
	Handle(GetWorkloadsParams) middleware.Responder
}

// NewGetWorkloads creates a new http.Handler for the get workloads operation
func NewGetWorkloads(ctx *middleware.Context, handler GetWorkloadsHandler) *GetWorkloads {
	return &GetWorkloads{Context: ctx, Handler: handler}
}

/* GetWorkloads swagger:route GET /wm/workloads workload-monitor getWorkloads

returns all workloads registered with the monitor

*/
type GetWorkloads struct {
	Context *middleware.Context
	Handler GetWorkloadsHandler
}

func (o *GetWorkloads) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetWorkloadsParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
