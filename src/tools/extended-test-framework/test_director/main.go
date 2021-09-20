package main

import (
	"fmt"
	"log"

	"os"
	"time"

	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"

	"mayastor-e2e/tools/extended-test-framework/test_director/restapi"
	"mayastor-e2e/tools/extended-test-framework/test_director/restapi/operations"

	"k8s.io/client-go/rest"
)

func banner() {
	fmt.Println("test_director started")
}

func startServer() {
	log.Printf("Server started 3")

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewEtfwAPI(swaggerSpec)
	server := restapi.NewServer(api)
	//defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Test Framework API"
	parser.LongDescription = "MayaData System Test Framework API"

	log.Printf("test_director about to configure flags")

	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
	log.Printf("test_director about to configure")
	server.ConfigureAPI()

	log.Printf("test_director about to serve")
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}

func main() {
	banner()

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get config")
		return
	}

	if restConfig == nil {
		fmt.Println("failed to get restConfig")
		return
	}

	go startServer()

	fmt.Println("waiting")
	time.Sleep(6000 * time.Second)
	fmt.Println("finishing")
}
