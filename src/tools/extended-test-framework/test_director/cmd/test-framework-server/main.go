// Code generated by go-swagger; DO NOT EDIT.

package main

import (
	"log"
	"os"
	"test-director/config"
	"test-director/handlers"
	"test-director/logger"
	"test-director/utils"

	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"

	"test-director/restapi"
	"test-director/restapi/operations"
)

// This file was generated by the swagger tool.
// Make sure not to overwrite this file after you generated it because all your edits would be lost!

func main() {

	configPath := utils.GetConfigPath("local")

	cfgFile, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("LoadConfig: %v", err)
	}

	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		log.Fatalf("ParseConfig: %v", err)
	}
	logger.InitLogger(&cfg.Logger)

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	handlers.InitEventCache()
	handlers.InitTestPlanCache(&cfg.Server)
	handlers.InitTestRunCache()

	api := operations.NewTestFrameworkAPI(swaggerSpec)
	server := restapi.NewServer(api)
	defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Test Framework"
	parser.LongDescription = "MayaData System Test Framework API"
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

	server.ConfigureAPI()

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}