// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"test-director/handlers"
	"test-director/restapi/operations"
	"test-director/restapi/operations/test_director"
)

//go:generate swagger generate server --target ../../test_director --name TestFramework --spec ../swagger_test_director_oas2.yaml --principal interface{}

func configureFlags(api *operations.TestFrameworkAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.TestFrameworkAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.TestDirectorGetTestPlansHandler = handlers.NewGetTestPlansHandler()
	api.TestDirectorGetTestPlanByIDHandler = handlers.NewGetTestPlanByIdHandler()
	api.TestDirectorDeleteTestPlanByIDHandler = handlers.NewDeleteTestPlanByIdHandler()
	api.TestDirectorDeleteTestPlansHandler = handlers.NewDeleteTestPlansHandler()
	api.TestDirectorPutTestPlanByIDHandler = handlers.NewPutTestPlanHandler()

	api.TestDirectorGetTestRunsHandler = handlers.NewGetTestRunsHandler()
	api.TestDirectorGetTestRunByIDHandler = handlers.NewGetTestRunByIdHandler()
	api.TestDirectorDeleteTestRunByIDHandler = handlers.NewDeleteTestRunByIdHandler()
	api.TestDirectorPutTestRunByIDHandler = handlers.NewPutTestRunHandler()

	api.TestDirectorAddEventHandler = handlers.NewAddEventHandler()
	api.TestDirectorGetEventsHandler = handlers.NewGetEventsHandler()

	if api.TestDirectorAddEventHandler == nil {
		api.TestDirectorAddEventHandler = test_director.AddEventHandlerFunc(func(params test_director.AddEventParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.AddEvent has not yet been implemented")
		})
	}
	if api.TestDirectorDeleteTestPlanByIDHandler == nil {
		api.TestDirectorDeleteTestPlanByIDHandler = test_director.DeleteTestPlanByIDHandlerFunc(func(params test_director.DeleteTestPlanByIDParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.DeleteTestPlanByID has not yet been implemented")
		})
	}
	if api.TestDirectorDeleteTestPlansHandler == nil {
		api.TestDirectorDeleteTestPlansHandler = test_director.DeleteTestPlansHandlerFunc(func(params test_director.DeleteTestPlansParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.DeleteTestPlans has not yet been implemented")
		})
	}
	if api.TestDirectorDeleteTestRunByIDHandler == nil {
		api.TestDirectorDeleteTestRunByIDHandler = test_director.DeleteTestRunByIDHandlerFunc(func(params test_director.DeleteTestRunByIDParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.DeleteTestRunByID has not yet been implemented")
		})
	}
	if api.TestDirectorGetEventsHandler == nil {
		api.TestDirectorGetEventsHandler = test_director.GetEventsHandlerFunc(func(params test_director.GetEventsParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.GetEvents has not yet been implemented")
		})
	}
	if api.TestDirectorGetTestPlanByIDHandler == nil {
		api.TestDirectorGetTestPlanByIDHandler = test_director.GetTestPlanByIDHandlerFunc(func(params test_director.GetTestPlanByIDParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.GetTestPlanByID has not yet been implemented")
		})
	}
	if api.TestDirectorGetTestPlansHandler == nil {
		api.TestDirectorGetTestPlansHandler = test_director.GetTestPlansHandlerFunc(func(params test_director.GetTestPlansParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.GetTestPlans has not yet been implemented")
		})
	}
	if api.TestDirectorGetTestRunByIDHandler == nil {
		api.TestDirectorGetTestRunByIDHandler = test_director.GetTestRunByIDHandlerFunc(func(params test_director.GetTestRunByIDParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.GetTestRunByID has not yet been implemented")
		})
	}
	if api.TestDirectorGetTestRunsHandler == nil {
		api.TestDirectorGetTestRunsHandler = test_director.GetTestRunsHandlerFunc(func(params test_director.GetTestRunsParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.GetTestRuns has not yet been implemented")
		})
	}
	if api.TestDirectorPutTestPlanByIDHandler == nil {
		api.TestDirectorPutTestPlanByIDHandler = test_director.PutTestPlanByIDHandlerFunc(func(params test_director.PutTestPlanByIDParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.PutTestPlanByID has not yet been implemented")
		})
	}
	if api.TestDirectorPutTestRunByIDHandler == nil {
		api.TestDirectorPutTestRunByIDHandler = test_director.PutTestRunByIDHandlerFunc(func(params test_director.PutTestRunByIDParams) middleware.Responder {
			return middleware.NotImplemented("operation test_director.PutTestRunByID has not yet been implemented")
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
	return addLogging(handler)
}

func addLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request:", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func addLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request:", r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}
