// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/handlers"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/restapi/operations"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/restapi/operations/workload_monitor"
)

//go:generate swagger generate server --target ../../workload_monitor --name Etfw --spec ../swagger_workload_monitor_oas2.yaml --principal interface{}

func configureFlags(api *operations.EtfwAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.EtfwAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.WorkloadMonitorPutWorkloadByRegistrantHandler = handlers.NewPutWorkloadByRegistrantHandler()
	api.WorkloadMonitorGetWorkloadByRegistrantHandler = handlers.NewGetWorkloadByRegistrantHandler()
	api.WorkloadMonitorDeleteWorkloadByRegistrantHandler = handlers.NewDeleteWorkloadByRegistrantHandler()
	api.WorkloadMonitorDeleteWorkloadsByRegistrantHandler = handlers.NewDeleteWorkloadsByRegistrantHandler()

	if api.WorkloadMonitorDeleteWorkloadByRegistrantHandler == nil {
		api.WorkloadMonitorDeleteWorkloadByRegistrantHandler = workload_monitor.DeleteWorkloadByRegistrantHandlerFunc(func(params workload_monitor.DeleteWorkloadByRegistrantParams) middleware.Responder {
			return middleware.NotImplemented("operation workload_monitor.DeleteWorkloadByRegistrant has not yet been implemented")
		})
	}
	if api.WorkloadMonitorDeleteWorkloadsByRegistrantHandler == nil {
		api.WorkloadMonitorDeleteWorkloadsByRegistrantHandler = workload_monitor.DeleteWorkloadsByRegistrantHandlerFunc(func(params workload_monitor.DeleteWorkloadsByRegistrantParams) middleware.Responder {
			return middleware.NotImplemented("operation workload_monitor.DeleteWorkloadsByRegistrant has not yet been implemented")
		})
	}
	if api.WorkloadMonitorGetWorkloadByRegistrantHandler == nil {
		api.WorkloadMonitorGetWorkloadByRegistrantHandler = workload_monitor.GetWorkloadByRegistrantHandlerFunc(func(params workload_monitor.GetWorkloadByRegistrantParams) middleware.Responder {
			return middleware.NotImplemented("operation workload_monitor.GetWorkloadByRegistrant has not yet been implemented")
		})
	}
	if api.WorkloadMonitorGetWorkloadsHandler == nil {
		api.WorkloadMonitorGetWorkloadsHandler = workload_monitor.GetWorkloadsHandlerFunc(func(params workload_monitor.GetWorkloadsParams) middleware.Responder {
			return middleware.NotImplemented("operation workload_monitor.GetWorkloads has not yet been implemented")
		})
	}
	if api.WorkloadMonitorGetWorkloadsByRegistrantHandler == nil {
		api.WorkloadMonitorGetWorkloadsByRegistrantHandler = workload_monitor.GetWorkloadsByRegistrantHandlerFunc(func(params workload_monitor.GetWorkloadsByRegistrantParams) middleware.Responder {
			return middleware.NotImplemented("operation workload_monitor.GetWorkloadsByRegistrant has not yet been implemented")
		})
	}
	if api.WorkloadMonitorPutWorkloadByRegistrantHandler == nil {
		api.WorkloadMonitorPutWorkloadByRegistrantHandler = workload_monitor.PutWorkloadByRegistrantHandlerFunc(func(params workload_monitor.PutWorkloadByRegistrantParams) middleware.Responder {
			return middleware.NotImplemented("operation workload_monitor.PutWorkloadByRegistrant has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
