package swagger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	swg "github.com/go-openapi/spec"
)

const (
	ProducesAllRegex = "\\*\\/\\*"
	ExampleRegex     = "\"example\":.*"

	AWSExtensionsFieldName = "x-amazon-apigateway-integration"

	ExpectedAuthorizerArn = "arn:aws:apigateway:eu-west-2:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-2:326458601802:function:api-gateway-authorizer-dev-auth/invocations"

	ExpectedAuthorizerName = "wave-api-gw-dev"

	// CheckMark used for unit test highlight.
	CheckMark = "\u2713"

	// BallotX used for unit test highlight.
	BallotX = "\u2717"
)

func TestSwaggerClient_RenderSwaggerShouldReplaceProuceAllStarWithApplicationJson(t *testing.T) {

	t.Logf("Given we read swagger from the deployed account service")
	{
		t.Logf("\tWhen calling RenderSwagger method, it should remove all the tags example")
		{
			// read vanilla swagger doc
			var data swg.Swagger
			swagger, _ := ioutil.ReadFile("../data/swagger.json")

			// check the vanilla swagger contains /*/

			regex := regexp.MustCompile(ProducesAllRegex)

			if regex.MatchString(string(swagger)) {
				t.Logf("\t\tVanilla swagger doc should have produces all type %v", CheckMark)
			} else {
				t.Errorf("\t\tVanilla swagger doc shoudl have produces all type %v", BallotX)
			}

			decoder := json.NewDecoder(bytes.NewReader(swagger))
			decoder.Decode(&data)

			// swagger client
			client := NewSwaggerClient("url")

			// render the swagger
			renderSwagger, _ := client.RenderSwagger(data)

			swagDoc := string(renderSwagger)

			if !regex.MatchString(swagDoc) {
				t.Logf("\t\tProduces all * should have been removed from the swagge doc %v", CheckMark)
			} else {
				t.Errorf("\t\tProduces all * should have been removed from the swagge doc %v", BallotX)
			}

			// Checks all Http error code are set properly
			for _, mappedError := range mappedErrors {
				if strings.Contains(swagDoc, fmt.Sprintf("\"statusCode\":\"%s\"", mappedError)) {
					t.Logf("\t\t Mappinng for error %s is set %v", mappedError, CheckMark)
				} else {
					t.Errorf("\t\tMappinng for error %s is set %v", mappedError, BallotX)
				}
			}
		}
	}

}

func TestSwaggerClient_RenderSwaggerMustAddAWSExtensions(t *testing.T) {

	t.Logf("Given we read swagger from the deployed account service")
	{
		t.Logf("\tWhen calling RenderSwagger method, it should have added the AWS extension")
		{
			// read vanilla swagger doc
			var data swg.Swagger
			swagger, _ := ioutil.ReadFile("../data/swagger.json")

			// check the vanilla swagger contains /*/

			if !strings.Contains(string(swagger), AWSExtensionsFieldName) {
				t.Logf("\t\tVanilla swagger must not have aws extensions  %v", CheckMark)
			} else {
				t.Errorf("\t\tVanilla swagger must not have aws extensions %v", BallotX)
			}

			decoder := json.NewDecoder(bytes.NewReader(swagger))
			decoder.Decode(&data)

			// swagger client
			client := NewSwaggerClient("account-service")

			// render the swagger
			renderSwagger, _ := client.RenderSwagger(data)

			if strings.Contains(string(renderSwagger), AWSExtensionsFieldName) {
				t.Logf("\t\tRendered swagger must have aws extensions  %v", CheckMark)
			} else {
				t.Errorf("\t\tRendered swagger must have aws extensions %v", BallotX)
			}

			logrus.Info(string(renderSwagger))
			if strings.Contains(string(renderSwagger), ExpectedAuthorizerArn) {
				t.Logf("\t\tRendered swagger must have the correct Authorizer Arn set  %v", CheckMark)
			} else {
				t.Errorf("\t\tRendered swagger must have the correct Authorizer Arn set %v", BallotX)
			}

			if strings.Contains(string(renderSwagger), ExpectedAuthorizerName) {
				t.Logf("\t\tRendered swagger must have the correct Authorizer name set  %v", CheckMark)
			} else {
				t.Errorf("\t\tRendered swagger must have the correct Authorizer name set %v", BallotX)
			}

		}
	}
}

func TestIsPathVisible_ReturnTrueForExplicitlyPublishedPathsAndNoExplicitlyPublished(t *testing.T) {

	t.Logf("Given we read swagger from the deployed account service")
	{
		t.Logf("\tWhen calling isPathVisible method, it should only return true when it's an explicitly published path")
		{
			// read vanilla swagger doc
			var data swg.Swagger
			swagger, _ := ioutil.ReadFile("../data/swagger.json")

			decoder := json.NewDecoder(bytes.NewReader(swagger))
			decoder.Decode(&data)

			var expectedPathVisibilityResults = map[string]bool{
				"/organisations/{orgId}":               true,
				"/accounts/{accountId}":                true,
				"/accounts/{accountId}/contacts/email": true,
				"/accounts/{accountId}/status":         true,
				"/admin/accounts":                      true,
				"/admin/accounts/email/{email}":        true,
				"/admin/accounts/{accountId}":          false,
			}

			for key, value := range data.Paths.Paths {
				if _, ok := expectedPathVisibilityResults[key]; !ok {
					t.Errorf("isPathVisible(%s) couldn't be called, unexpected path not found in result table", key)
					continue
				}
				res := isPathVisible(value)

				if res != expectedPathVisibilityResults[key] {
					t.Errorf("isPathVisible(%s) was wrong, expected %t, got %t", key, expectedPathVisibilityResults[key], res)
				}
			}

		}
	}
}

func TestSwaggerClient_RenderSwaggerShouldRemoveAllExamplesTag(t *testing.T) {

	t.Logf("Given we read swagger from the deployed account service")
	{
		t.Logf("\tWhen calling RenderSwagger method, it should remove all the example tags")
		{
			// read vanilla swagger doc
			var data swg.Swagger
			swagger, _ := ioutil.ReadFile("../data/swagger.json")

			// check the vanilla swagger contains /*/

			regex := regexp.MustCompile(ExampleRegex)

			if regex.MatchString(string(swagger)) {
				t.Logf("\t\tVanilla swagger doc should have example tags %v", CheckMark)
			} else {
				t.Errorf("\t\tVanilla swagger doc should have example tags %v", BallotX)
			}

			decoder := json.NewDecoder(bytes.NewReader(swagger))
			decoder.Decode(&data)

			// swagger client
			client := NewSwaggerClient("account-service")

			// render the swagger
			renderSwagger, _ := client.RenderSwagger(data)

			if !regex.MatchString(string(renderSwagger)) {
				t.Logf("\t\tRendered swagger do should not have example tag %v", CheckMark)
			} else {
				t.Errorf("\t\tRendered swagger do should not have example tag %v", BallotX)
			}
		}
	}

}

func TestSwaggerClient_RenderSwaggerShouldRenameNonAplphanumericEntityModel(t *testing.T) {

	t.Logf("Given we read swagger from the deployed account service")
	{
		t.Logf("\tWhen calling RenderSwagger method, it should remove all the nonalphanumeric entity model")
		{
			// read vanilla swagger doc
			var data swg.Swagger
			swagger, _ := ioutil.ReadFile("../data/swagger-nonalphanumeric.json")

			regex := regexp.MustCompile(GoModelRegex)

			if regex.MatchString(string(swagger)) {
				t.Logf("\t\tVanilla swagger doc should have nonalphanumeric entity model defined %v", CheckMark)
			} else {
				t.Errorf("\t\tVanilla swagger doc should have nonalphanumeric entity model defined %v", BallotX)
			}

			decoder := json.NewDecoder(bytes.NewReader(swagger))
			decoder.Decode(&data)

			// swagger client
			client := NewSwaggerClient("account-service")

			// render the swagger
			renderSwagger, _ := client.RenderSwagger(data)

			if !regex.MatchString(string(renderSwagger)) {
				t.Logf("\t\tRendered swagger doc should not have nonalphanumeric entity model defined %v", CheckMark)
			} else {
				t.Errorf("\t\tRendered swagger doc should not have nonalphanumeric entity model defined  %v", BallotX)
			}
		}
	}
}

func TestAddCORSSupport(t *testing.T) {
	t.Run("Should add an options operation to every endpoint", func(t *testing.T) {
		// read vanilla swagger doc

		os.Setenv(ApiGwName, "api-gw-dev")
		var data swg.Swagger
		swagger, _ := ioutil.ReadFile("../data/swagger.json")

		decoder := json.NewDecoder(bytes.NewReader(swagger))
		decoder.Decode(&data)

		// swagger client
		client := NewSwaggerClient("account-service")

		// render the swagger
		renderSwagger, _ := client.RenderSwagger(data)

		var dataResult swg.Swagger
		decoderResult := json.NewDecoder(bytes.NewReader(renderSwagger))
		decoderResult.Decode(&dataResult)

		for key, path := range dataResult.Paths.Paths {
			if path.Options != nil {
				t.Logf("\t\tEndpoint [%s] has Options Operation %v", key, CheckMark)
			} else {
				t.Errorf("\t\tEndpoint [%s] does not have expected Options Operation %v", key, BallotX)
			}
		}
	})

	t.Run("Options Operation Should return Status:200", func(t *testing.T) {
		// read vanilla swagger doc
		var data swg.Swagger
		swagger, _ := ioutil.ReadFile("../data/swagger.json")

		decoder := json.NewDecoder(bytes.NewReader(swagger))
		decoder.Decode(&data)

		// swagger client
		client := NewSwaggerClient("account-service")

		// render the swagger
		renderSwagger, _ := client.RenderSwagger(data)

		var dataResult swg.Swagger
		decoderResult := json.NewDecoder(bytes.NewReader(renderSwagger))
		decoderResult.Decode(&dataResult)

		for key, path := range dataResult.Paths.Paths {
			options := path.Options

			_, ok := options.Responses.ResponsesProps.StatusCodeResponses[200]

			if ok {
				t.Logf("\t\tEndpoint [%s] has Options:200 Response Entry %v", key, CheckMark)
			} else {
				t.Errorf("\t\tEndpoint [%s] does not have Options:200 Response Entry %v", key, BallotX)
			}
		}
	})

	t.Run("Options Operation Should return expected headers under status 200", func(t *testing.T) {
		// read vanilla swagger doc
		var data swg.Swagger
		swagger, _ := ioutil.ReadFile("../data/swagger.json")

		decoder := json.NewDecoder(bytes.NewReader(swagger))
		decoder.Decode(&data)

		// swagger client
		client := NewSwaggerClient("account-service")

		// render the swagger
		renderSwagger, _ := client.RenderSwagger(data)

		var dataResult swg.Swagger
		decoderResult := json.NewDecoder(bytes.NewReader(renderSwagger))
		decoderResult.Decode(&dataResult)

		expectedHeaders := []string{
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Methods",
			"Access-Control-Allow-Headers",
		}

		for key, path := range dataResult.Paths.Paths {
			options := path.Options

			successCode, _ := options.Responses.ResponsesProps.StatusCodeResponses[200]

			for _, expected := range expectedHeaders {
				_, ok := successCode.ResponseProps.Headers[expected]

				if ok {
					t.Logf("\t\tExpected Header [%s] present for endpoint [%s] %v",
						expected, key, CheckMark)
				} else {
					t.Logf("\t\tExpected Heaader [%s] is not present for endpoint [%s] %v",
						expected, key, BallotX)
				}
			}
		}
	})

	t.Run("Should add an 'Access-Control-Allow-Origin' header to every endpoint, operation, and response", func(t *testing.T) {
		// read vanilla swagger doc
		var data swg.Swagger
		swagger, _ := ioutil.ReadFile("../data/swagger.json")

		decoder := json.NewDecoder(bytes.NewReader(swagger))
		decoder.Decode(&data)

		// swagger client
		client := NewSwaggerClient("account-service")

		// render the swagger
		renderSwagger, _ := client.RenderSwagger(data)

		var dataResult swg.Swagger
		decoderResult := json.NewDecoder(bytes.NewReader(renderSwagger))
		decoderResult.Decode(&dataResult)

		accessControlAllowOriginHeader := "Access-Control-Allow-Origin"

		for key, path := range dataResult.Paths.Paths {
			if path.Get != nil {
				for statusCode, response := range path.Get.OperationProps.Responses.ResponsesProps.StatusCodeResponses {
					_, ok := response.ResponseProps.Headers[accessControlAllowOriginHeader]
					if !ok {
						t.Errorf("\t\tStatus Code [%d] for GET endpoint [%s] does not have %s Header %v",
							statusCode, key, accessControlAllowOriginHeader, BallotX)
					}
				}
			}
			if path.Patch != nil {
				for statusCode, response := range path.Patch.OperationProps.Responses.ResponsesProps.StatusCodeResponses {
					_, ok := response.ResponseProps.Headers[accessControlAllowOriginHeader]
					if !ok {
						t.Errorf("\t\tStatus Code [%d] for PATCH endpoint [%s] does not have %s Header %v",
							statusCode, key, accessControlAllowOriginHeader, BallotX)
					}
				}
			}
			if path.Post != nil {
				for statusCode, response := range path.Post.OperationProps.Responses.ResponsesProps.StatusCodeResponses {
					_, ok := response.ResponseProps.Headers[accessControlAllowOriginHeader]
					if !ok {
						t.Errorf("\t\tStatus Code [%d] for POST endpoint [%s] does not have %s Header %v",
							statusCode, key, accessControlAllowOriginHeader, BallotX)
					}
				}
			}
			if path.Put != nil {
				for statusCode, response := range path.Put.OperationProps.Responses.ResponsesProps.StatusCodeResponses {
					_, ok := response.ResponseProps.Headers[accessControlAllowOriginHeader]
					if !ok {
						t.Errorf("\t\tStatus Code [%d] for PUT endpoint [%s] does not have %s Header %v",
							statusCode, key, accessControlAllowOriginHeader, BallotX)
					}
				}
			}
			if path.Delete != nil {
				for statusCode, response := range path.Delete.OperationProps.Responses.ResponsesProps.StatusCodeResponses {
					_, ok := response.ResponseProps.Headers[accessControlAllowOriginHeader]
					if !ok {
						t.Errorf("\t\tStatus Code [%d] for DELETE endpoint [%s] does not have %s Header %v",
							statusCode, key, accessControlAllowOriginHeader, BallotX)
					}
				}
			}
		}
	})
}

func TestContentTypeTemplate_ParserShouldLookForConsumeTypeToSetTheCorrectContentTypeTemplate(t *testing.T) {

	t.Logf("Given we read swagger from the deployed account service")
	{
		t.Logf("\tWhen rendering the swagger doc, the appropriate content type template is set for each endpoint")
		{
			// read vanilla swagger doc
			var data swg.Swagger
			swagger, _ := ioutil.ReadFile("../data/swagger_multipart_type.json")

			decoder := json.NewDecoder(bytes.NewReader(swagger))
			decoder.Decode(&data)

			client := NewSwaggerClient("ocr-api-service")
			renderSwagger, _ := client.RenderSwagger(data)

			fmt.Println(string(renderSwagger))

			var dataResult swg.Swagger
			decoderResult := json.NewDecoder(bytes.NewReader(renderSwagger))
			decoderResult.Decode(&dataResult)

			pathItem := dataResult.Paths.Paths["/ocrs"]

			postOperation := pathItem.Post
			getOperation := pathItem.Get

			// Post
			authHeader := getRequestParameterValue(postOperation, "integration.request.header.X-JWT-Assertion")
			orgIdPosHeader := getRequestParameterValue(getOperation, "integration.request.header.organisation-id")
			contentTypePostHeader := getRequestParameterValue(getOperation, "integration.request.header.content-type")
			acceptPostHeader := getRequestParameterValue(getOperation, "integration.request.header.accept")

			assertHeader(authHeader, "context.authorizer.stringKey", t)
			assertHeader(orgIdPosHeader, "method.request.header.organisation-id", t)
			assertHeader(contentTypePostHeader, "method.request.header.content-type", t)
			assertHeader(acceptPostHeader, "method.request.header.accept", t)

			// Get
			authGetHeader := getRequestParameterValue(getOperation, "integration.request.header.X-JWT-Assertion")
			orgIdGetHeader := getRequestParameterValue(getOperation, "integration.request.header.organisation-id")
			contentTypeGetHeader := getRequestParameterValue(getOperation, "integration.request.header.content-type")
			acceptGetHeader := getRequestParameterValue(getOperation, "integration.request.header.accept")

			assertHeader(authGetHeader, "context.authorizer.stringKey", t)
			assertHeader(orgIdGetHeader, "method.request.header.organisation-id", t)
			assertHeader(contentTypeGetHeader, "method.request.header.content-type", t)
			assertHeader(acceptGetHeader, "method.request.header.accept", t)
		}
	}
}

func TestSwaggerIncludesOnfidoHeader(t *testing.T) {
	t.Run("Should include custom integration header key and value", func(t *testing.T) {
		// read vanilla swagger doc
		var data swg.Swagger
		swagger, _ := ioutil.ReadFile("../data/swagger_onfido_header.json")

		decoder := json.NewDecoder(bytes.NewReader(swagger))
		decoder.Decode(&data)

		// swagger client
		client := NewSwaggerClient("verification-service")

		// render the swagger
		renderSwagger, _ := client.RenderSwagger(data)

		// Expected rendered strings
		expectedStrings := []string{
			"integration.request.header.X-Signature",
			"method.request.header.X-Signature"}

		for _, val := range expectedStrings {
			regex := regexp.MustCompile(val)

			if regex.MatchString(string(renderSwagger)) {
				t.Logf("\t\tRendered swagger contains expected value [%s] %v", val, CheckMark)
			} else {
				t.Errorf("\t\tRendered swagger does not have value [%s] %v", val, BallotX)
			}
		}
	})
}

func TestSwaggerIncludesDeclaredHeaders(t *testing.T) {
	t.Run("Should include custom integration header key and value", func(t *testing.T) {
		// read vanilla swagger doc
		var data swg.Swagger
		swagger, _ := ioutil.ReadFile("../data/swagger_header.json")

		decoder := json.NewDecoder(bytes.NewReader(swagger))
		decoder.Decode(&data)

		// swagger client
		client := NewSwaggerClient("receipt-bank-processor")

		// render the swagger
		renderSwagger, _ := client.RenderSwagger(data)

		// Expected rendered strings
		expectedStrings := []string{
			"integration.request.header.X-Rb-Funnel-Api-Key",
			"method.request.header.X-Rb-Funnel-Api-Key"}

		for _, val := range expectedStrings {
			regex := regexp.MustCompile(val)

			if regex.MatchString(string(renderSwagger)) {
				t.Logf("\t\tRendered swagger contains expected value [%s] %v", val, CheckMark)
			} else {
				t.Errorf("\t\tRendered swagger contains expected value [%s] %v", val, BallotX)
			}
		}
	})
}

func assertHeader(authHeader, expectedValue string, t *testing.T) {
	if authHeader == expectedValue {
		t.Logf("\t\t The request parameter for auth header should be set to %v %v", expectedValue, CheckMark)
	} else {
		t.Errorf("\t\t The request parameter for auth header should be set to %v %v", expectedValue, BallotX)
	}
}

// fetches the template value for given content type
func getRequestParameterValue(op *swg.Operation, headerName string) string {
	extension := op.Extensions["x-amazon-apigateway-integration"]
	xAmazonIntegrations := extension.(map[string]interface{})
	requestParameters := xAmazonIntegrations["requestParameters"].(map[string]interface{})
	return requestParameters[headerName].(string)
}
