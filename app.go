package main

import (
	"github.com/akhettar/apigw-pub/apigw"
	"github.com/akhettar/apigw-pub/swagger"
	"github.com/akhettar/apigw-pub/utils"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	StageNameVarKey = "STAGE_NAME"
	APIGatewayIDKey = "API_GATEWAY_ID"
	SwaggerUrl      = "SWAGGER_URL"
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
	client := swagger.NewSwaggerClient(utils.RetrieveEnvVar(SwaggerUrl))
	doc, err := client.FetchSwagger()
	if err != nil {
		log.WithFields(log.Fields{"Swagger Url": utils.RetrieveEnvVar(SwaggerUrl)}).Fatal("Failed to retrieve swagger document")
	}
	renderedSwag, err := client.RenderSwagger(doc)
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
}
