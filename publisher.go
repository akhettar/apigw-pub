package main

import (
	flags "github.com/jessevdk/go-flags"
	"os"

	"github.com/akhettar/apigw-pub/apigw"
	"github.com/akhettar/apigw-pub/swagger"
	"github.com/akhettar/apigw-pub/utils"
	log "github.com/sirupsen/logrus"
)


var opts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	SwaggerURL []bool `short:"swg_url" long:"swag_url" description:"the swagger url" required:"true"`

	// Example of automatic marshalling to desired type (uint)
	Offset uint `long:"offset" description:"Offset"`

	// Example of a callback, called each time the option is found.
	Call func(string) `short:"c" description:"Call phone number"`

	// Example of a required flag
	Name string `short:"n" long:"name" description:"A name" required:"true"`

	// Example of a flag restricted to a pre-defined set of strings
	Animal string `long:"animal" choice:"cat" choice:"dog"`
}
func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)


	
}

const (
	StageNameVarKey = "STAGE_NAME"
	APIGatewayIDKey = "API_GATEWAY_ID"
	SwaggerUrl      = "SWAGGER_URL"
)

// Publisher the swagger publisher type
type Publisher struct {
	parser swagger.SwaggerParser
	client apigw.APIGatewayClient
}

// NewPublisher create a new instance of the publisher
func NewPublisher() *Publisher {

	_, err := flags.ParseArgs(&opts)

if err != nil {
	panic(err)
}

	url := flag.String("swagger_url", "", "The url of the swagger document that can be sourced from the live running server")
	parser := swagger.NewSwaggerClient(utils.RetrieveEnvVar(*url))

	doc, err := parser.FetchSwagger()
	if err != nil {
		log.WithFields(log.Fields{"Swagger Url": utils.RetrieveEnvVar(SwaggerUrl)}).Fatal("Failed to retrieve swagger document")
	}
	renderedSwag, err := parser.RenderSwagger(doc)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to render swagger document")
	}

	// Import swagger
	apigwClient := apigw.NewAPIGatewayClient()
	report, err := apigwClient.ImportSwagger(renderedSwag, utils.RetrieveEnvVar(APIGatewayIDKey))

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to publish the swagger doc")
	}
	log.Info(report)

	// Deploy API
	deployment, err := apigwClient.CreateDeployment(utils.RetrieveEnvVar(StageNameVarKey),
		utils.RetrieveEnvVar(APIGatewayIDKey))
	log.Info(deployment)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to deploy the newly created resources  ❌")
	}
	log.Info("Swagger import and deployment is successfully completed ✅")

	return &Publisher{parser, apigwClient}
}

func (pub *Publisher) run() {
	doc, err := pub.parser.FetchSwagger()
	if err != nil {
		log.WithFields(log.Fields{"Swagger Url": utils.RetrieveEnvVar(SwaggerUrl)}).Fatal("Failed to retrieve swagger document")
	}
	renderedSwag, err := pub.parser.RenderSwagger(doc)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to render swagger document")
	}

	// Import swagger
	
	report, err := pub.client.ImportSwagger(renderedSwag, utils.RetrieveEnvVar(APIGatewayIDKey))

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to publish the swagger doc")
	}
	log.Info(report)

	// Deploy API
	deployment, err := pub.client.CreateDeployment(utils.RetrieveEnvVar(StageNameVarKey),
		utils.RetrieveEnvVar(APIGatewayIDKey))
	log.Info(deployment)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Fatal("Failed to deploy the newly created resources  ❌")
	}
	log.Info("Swagger import and deployment is successfully completed ✅")

}
