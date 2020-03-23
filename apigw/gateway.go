package apigw

import (
	"github.com/akhettar/aws-apigw-publisher/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	AssumeRole = "ASSUME_ROLE"
	Region     = "AWS_REGION"
)

type APIGatewayClient struct {
	apigw *apigateway.APIGateway
}

// SDK Client
// Initial credentials loaded from SDK's default credential chain. Such as
// the environment, shared credentials (~/.aws/credentials), or EC2 Instance
// The assume role is used when run locally against the dev environment
func NewAPIGatewayClient() APIGatewayClient {

	// Session
	ses := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(utils.FetchEnvVar(Region, endpoints.EuWest1RegionID)),
		Credentials: credentials.NewEnvCredentials(),
	}))

	if urn, ok := os.LookupEnv(AssumeRole); ok {
		log.WithFields(log.Fields{}).Info("Running with assuming role: ", urn)
		return createClientWithAssumeRole(ses, urn)

	}
	// running with aws iam user which has permission to publish to api gateway
	return createClientWithDefaultCredentials(ses)
}

// ImportSwagger imports the swagger doc into API Gateway
func (cl APIGatewayClient) ImportSwagger(swaggerDoc []byte, apigwId string) (*apigateway.RestApi, error) {
	log.WithFields(log.Fields{"API Gateway": apigwId}).Info("Importing swagger")
	put := apigateway.PutRestApiInput{RestApiId: &apigwId, Body: swaggerDoc}
	return cl.apigw.PutRestApi(&put)
}

// CreateDeployment for the recent upload
func (cl APIGatewayClient) CreateDeployment(stage string, apigwId string) (*apigateway.Deployment, error) {
	log.WithFields(log.Fields{"stage": stage, "API GatewayId": apigwId}).Info("Deploying API")
	createDep := apigateway.CreateDeploymentInput{RestApiId: &apigwId, StageName: &stage}
	return cl.apigw.CreateDeployment(&createDep)
}

// Creates a client with assume role - uses this when the toll is run locally (or when role in another AWS account needs to be assumed)
func createClientWithAssumeRole(sess *session.Session, urn string) APIGatewayClient {
	cred := stscreds.NewCredentials(sess, urn)
	return APIGatewayClient{apigateway.New(sess, &aws.Config{Credentials: cred})}
}

// Creates a client with IAM credentials only
func createClientWithDefaultCredentials(sess *session.Session) APIGatewayClient {
	return APIGatewayClient{apigateway.New(sess)}
}
