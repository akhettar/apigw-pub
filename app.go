package main

import (
	"encoding/json"
	"github.com/akhettar/aws-apigw-publisher/apigw"
	"github.com/akhettar/aws-apigw-publisher/swagger"
	"github.com/akhettar/aws-apigw-publisher/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	StageNameVarKey = "STAGE_NAME"
	APIGatewayIDKey = "API_GATEWAY_ID"
	ServiceName     = "SERVICE_NAME"
)

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

func main() {

	// the service name should be given as env variable
	servicename := utils.FetchEnvVar(ServiceName, "pet-store")

	log.WithFields(log.Fields{"Service name": servicename}).Info("Starting Swagger import and deployment for")

	// Fetching and rendering swagger
	client := swagger.NewSwaggerClient(servicename)
	swaggerDoc, err := client.FetchSwagger()

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalf("Failed to fetch or render the swagger doc")
	}

	// render the swagger - remove unsupported features by aws and add aws extensions
	renderedSwag, err := client.RenderSwagger(swaggerDoc)

	if err != nil {
		log.Errorln("Failed to render the swagger doc")
		log.Panic(err)
	}

	// Import swagger
	apigwClient := apigw.NewAPIGatewayClient()
	api, err := apigwClient.ImportSwagger(renderedSwag, os.Getenv(APIGatewayIDKey))

	if err != nil {
		log.WithFields(log.Fields{}).Errorf("Failed to publish the swagger doc")
		log.Panic(err)
	}
	prettyJSON, err := json.MarshalIndent(api, "", "         ")
	if err != nil {
		log.Fatal("Failed to generate json", err)
	}
	log.Info(string(prettyJSON))

	// Deploy API
	deployment, err := apigwClient.CreateDeployment(os.Getenv(StageNameVarKey), os.Getenv(APIGatewayIDKey))
	log.Info(deployment)
	if err != nil {
		log.Errorf("Failed to deploy the newly created resources  ❌")
		log.Panic(err)
	}
	log.WithFields(log.Fields{"Service name": servicename}).Info("Swagger import and deployment has been successfully completed ✅")
}
