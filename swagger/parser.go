package swagger

import (
	"encoding/json"
	"fmt"
	"github.com/akhettar/aws-apigw-publisher/model"
	"github.com/akhettar/aws-apigw-publisher/utils"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	swg "github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
)

var DEFAULT_JSON_MIME_TYPE = []string{"application/json"}

const (
	PublicConnectionType = "PUBLIC"
	ConnectionType       = "CONNECTION_TYPE"
	VPCLinkID            = "VPC_LINK_ID"
	GoModelRegex         = "(model.)(.*)"
	AuthType             = "AUTH_TYPE"
	AuthUrl              = "AUTH_URL"
	AuthName             = "AUTH_NAME"
	ApiGwName            = "API_GATEWAY_NAME"
	CustomAuth           = "apiKey"
	OAuth2               = "oauth2"
	EndpointUrl          = "ENDPOINT_URL"
	CorsEnabled          = "CORS_ENABLED"
	CustomHeaders        = "CUSTOM_HEADERS"
)

var alphaNumRegexp *regexp.Regexp
var mappedErrors [11]string
var contentTemplate map[string]string

func init() {
	alphaNumRegexp = regexp.MustCompile(GoModelRegex)
	mappedErrors = [11]string{"200", "201", "202", "204", "400", "401", "403", "404", "409", "424", "500"}
}

// SwaggerParser responsible for the followings:
// 1. Fetches the swagger document for a given service and a given environment
// 2. Apply filters such as removing `example` tag and some unsupported feature of OpenAPI by AWS REST API Gateway restrictions
// 3. Add AWS API Gateway integration extensions to the vanilla swagger doc
type SwaggerParser struct {
	swaggerUrl string
}

// NewSwaggerClient - Function
func NewSwaggerClient(swaggerUrl string) SwaggerParser {
	return SwaggerParser{swaggerUrl: swaggerUrl}
}

// FetchSwagger - Function
// Fetches Swagger for a given service from the its deployed environment
func (client SwaggerParser) FetchSwagger() (swg.Swagger, error) {

	log.WithFields(log.Fields{"Swagger URL": client.swaggerUrl}).Info("Fetching vanilla swagger from the given swagger url")

	resp, err := http.Get(client.swaggerUrl)
	if err != nil {
		log.Errorf("Error when getting Swagger docs: %s", err)
	}

	defer resp.Body.Close()

	var data swg.Swagger

	if err != nil {
		log.Errorf("Failed to fetch swagger doc from %s", client.swaggerUrl)
		return data, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Got error from the server with http code %d", resp.StatusCode)
		return data, fmt.Errorf("Got error from the server with http code %d", resp.StatusCode)
	}

	// parsing the swagger doc
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	return data, err
}

// RenderSwagger - Function
// Renders the vanilla swagger document into one that can be published to AWS api gateway
func (client SwaggerParser) RenderSwagger(doc swg.Swagger) ([]byte, error) {

	endpointUrl := utils.FetchEnvVar(EndpointUrl, fmt.Sprintf("%s%s", doc.Host, doc.BasePath))

	swaggerWithExtensions := swg.Swagger{
		SwaggerProps: doc.SwaggerProps,
	}

	// Setting the swagger Tile to that of the asto api gateway to avoid the overriding of the api gateway name by the REST API import call
	swaggerWithExtensions.Info.Title = utils.RetrieveEnvVar(ApiGwName)
	for key, value := range doc.Paths.Paths {
		if isPathVisible(value) {
			log.WithFields(log.Fields{"key": key}).Info(" Publishing ✅")
		} else {
			log.WithFields(log.Fields{"key": key}).Info(" Skipping Publish ❌")
			delete(swaggerWithExtensions.Paths.Paths, key)
		}
	}

	// Add custom authorization if set
	if strings.ToLower(os.Getenv(AuthType)) == strings.ToLower(CustomAuth) {
		swaggerWithExtensions.SecurityDefinitions = buildCustomAuthorizerBlock()
	}

	// Apply filters
	applyFilters(&swaggerWithExtensions)

	// adding aws extension for all the defined operations for a given endpoint
	for key, path := range doc.Paths.Paths {
		if path.Get != nil {
			addAWSExtensions(path.Get, key, http.MethodGet, endpointUrl, isSecurityEnabled(path))
			addOperationCORSHeaders(path.Get)
			renameNonAlphanumericReference(path.Get)
		}

		if path.Put != nil {
			addAWSExtensions(path.Put, key, http.MethodPut, endpointUrl, isSecurityEnabled(path))
			addOperationCORSHeaders(path.Put)
			renameNonAlphanumericReference(path.Put)
		}

		if path.Delete != nil {
			addAWSExtensions(path.Delete, key, http.MethodDelete, endpointUrl, isSecurityEnabled(path))
			addOperationCORSHeaders(path.Delete)
			renameNonAlphanumericReference(path.Delete)
		}

		if path.Patch != nil {
			addAWSExtensions(path.Patch, key, http.MethodPatch, endpointUrl, isSecurityEnabled(path))
			addOperationCORSHeaders(path.Patch)
			renameNonAlphanumericReference(path.Patch)
		}

		if path.Post != nil {
			addAWSExtensions(path.Post, key, http.MethodPost, endpointUrl, isSecurityEnabled(path))
			addOperationCORSHeaders(path.Post)
			renameNonAlphanumericReference(path.Post)
		}

		// cors enabled?
		_, corsEnabled := os.LookupEnv(CorsEnabled)
		if corsEnabled {
			pathPointer := doc.Paths.Paths[key]
			pathPointer.Options = swg.NewOperation("add_cors")
			addOptionsCORSSupport(pathPointer.Options, key, http.MethodOptions)
			doc.Paths.Paths[key] = pathPointer
		}
	}
	json, err := swaggerWithExtensions.MarshalJSON()
	return json, err
}

func buildCustomAuthorizerBlock() map[string]*swg.SecurityScheme {
	secWith := utils.RetrieveEnvVar(AuthName)
	return map[string]*swg.SecurityScheme{
		secWith: {SecuritySchemeProps: swg.SecuritySchemeProps{
			Type: "apiKey",
			Name: "Authorization",
			In:   "header",
		},
			VendorExtensible: swg.VendorExtensible{
				Extensions: map[string]interface{}{
					"x-amazon-apigateway-authtype": "custom",
					"x-amazon-apigateway-authorizer": map[string]interface{}{
						"authorizerUri":                utils.RetrieveEnvVar(AuthUrl),
						"authorizerResultTtlInSeconds": 0,
						"type":                         "token",
					},
				},
			},
		},
	}
}

func renameNonAlphanumericReference(operation *swg.Operation) {
	for _, response := range operation.Responses.ResponsesProps.StatusCodeResponses {
		if response.ResponseProps.Schema != nil {
			ref := response.ResponseProps.Schema.SchemaProps.Ref.Ref
			url := ref.GetURL()
			if url != nil && alphaNumRegexp.MatchString(url.Fragment) {
				url.Fragment = strings.ReplaceAll(url.Fragment, "model.", "")
				fmt.Println(url.Fragment)
			}
		}
	}

	// handling object parameters
	for _, param := range operation.Parameters {
		if param.Schema != nil && param.Schema.Ref.GetURL() != nil && alphaNumRegexp.MatchString(param.Schema.Ref.GetURL().Fragment) {
			param.Schema.Ref.GetURL().Fragment = strings.ReplaceAll(param.Schema.Ref.GetURL().Fragment, "model.", "")
		}
	}
}

// Adds Swagger Extensions
func addAWSExtensions(op *swg.Operation, key string, method string, endpointUrl string, securityEnabled bool) {
	requestParams := make(map[string]string)
	for _, param := range op.Parameters {
		if param.In == "path" {
			requestParams[fmt.Sprintf("integration.request.%s.%s", param.In, param.Name)] =
				fmt.Sprintf("method.request.%s.%s", param.In, param.Name)
		}
		if param.In == "query" {
			requestParams[fmt.Sprintf("integration.request.%s.%s", "querystring", param.Name)] =
				fmt.Sprintf("method.request.%s.%s", "querystring", param.Name)
		}
		if param.In == "header" {
			requestParams[fmt.Sprintf("integration.request.header.%s", param.Name)] =
				fmt.Sprintf("method.request.header.%s", param.Name)
		}
	}

	// set all the headers
	customHeaders, ok := os.LookupEnv(CustomHeaders)
	if ok {
		for _, header := range strings.Split(customHeaders, ",") {
			name := strings.TrimSpace(header)
			addHeaderParameter(op, name, "header", true, name)
			requestParams[fmt.Sprintf("integration.request.header.%s", name)] =
				fmt.Sprintf("method.request.header.%s", name)
		}
	}

	// set default headers
	addHeaderParameter(op, "content-type", "header", false, "content type")
	addHeaderParameter(op, "accept", "header", false, "accept")

	// set all the request parameters
	requestParams["integration.request.header.accept"] = "method.request.header.accept"
	requestParams["integration.request.header.content-type"] = "method.request.header.content-type"

	var responses = map[string]map[string]interface{}{}

	for _, element := range mappedErrors {
		responses[element] = map[string]interface{}{}
		responses[element]["statusCode"] = element
		responses[element]["responseParameters"] = map[string]string{
			"method.response.header.Access-Control-Allow-Origin": "'*'",
		}
	}

	op.Produces = DEFAULT_JSON_MIME_TYPE
	log.WithFields(log.Fields{
		"Endpoint": key,
	}).Info("Processing endpoint")

	extension := model.AWSAPIGatewayIntegration{
		ConnectionType:      strings.ToUpper(utils.FetchEnvVar(ConnectionType, PublicConnectionType)),
		URI:                 fmt.Sprintf("http://%s%s", endpointUrl, key),
		ConnectionID:        utils.FetchEnvVar(VPCLinkID, ""),
		HTTPMethod:          method,
		IntegrationType:     "http",
		PassthroughBehavior: "when_no_templates",
		RequestParameters:   requestParams,
		Responses:           responses,
	}

	item := op
	item.VendorExtensible.AddExtension("x-amazon-apigateway-integration", extension)

	if securityEnabled {
		item.SecuredWith(utils.RetrieveEnvVar(AuthName))
	} else {
		log.WithFields(log.Fields{
			"Endpoint": key,
		}).Warn("is marked as having no required authentication")
	}
}

// Add parameter to the given path
func addHeaderParameter(op *swg.Operation, paramName string, in string, required bool, description string) {
	parameters := op.Parameters
	schema := swg.SimpleSchema{Type: "string"}
	parametersProps := swg.ParamProps{Description: description, Name: paramName, In: in, Required: required}
	param := swg.Parameter{SimpleSchema: schema, ParamProps: parametersProps}
	op.Parameters = append(parameters, param)
}

func addOperationCORSHeaders(op *swg.Operation) {
	corsHeader := map[string]swg.Header{"Access-Control-Allow-Origin": {
		SimpleSchema: swg.SimpleSchema{Type: "string"},
	}}

	for key, response := range op.OperationProps.Responses.ResponsesProps.StatusCodeResponses {
		response.ResponseProps = swg.ResponseProps{
			Headers: corsHeader,
		}
		op.OperationProps.Responses.ResponsesProps.StatusCodeResponses[key] = response
	}
}

func addOptionsCORSSupport(op *swg.Operation, key, method string) {
	log.WithFields(log.Fields{"URL": key}).Info("Adding CORS Support to endpoint")

	op.OperationProps = swg.OperationProps{
		Summary:     "CORS Support",
		Description: "Enable CORS Support by returning correct headers",
		Consumes:    []string{"text/json", "application/json"},
		Produces:    []string{"text/json", "application/json"},
		Responses: &swg.Responses{
			VendorExtensible: swg.VendorExtensible{},
			ResponsesProps: swg.ResponsesProps{
				StatusCodeResponses: map[int]swg.Response{
					200: {
						Refable:          swg.Refable{},
						VendorExtensible: swg.VendorExtensible{},
						ResponseProps: swg.ResponseProps{
							Headers: map[string]swg.Header{
								"Access-Control-Allow-Origin": {
									SimpleSchema: swg.SimpleSchema{Type: "string"},
								},
								"Access-Control-Allow-Methods": {
									SimpleSchema: swg.SimpleSchema{Type: "string"},
								},
								"Access-Control-Allow-Headers": {
									SimpleSchema: swg.SimpleSchema{Type: "string"},
								},
							},
						},
					},
				},
			},
		},
	}

	extension := model.AWSAPIGatewayIntegration{
		IntegrationType:     "mock",
		PassthroughBehavior: "when_no_match",
		HTTPMethod:          method,
		RequestTemplates:    map[string]string{"application/json": "{\"statusCode\": 200}"},
		Responses: map[string]map[string]interface{}{
			"default": {
				"statusCode": "200",
				"responseParameters": map[string]string{
					"method.response.header.Access-Control-Allow-Methods": "'GET,OPTIONS,PATCH,PUT,POST,DELETE'",
					"method.response.header.Access-Control-Allow-Headers": "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,DNT,Origin,Referer,Sec-Fetch-Mode,User-Agent,Access-Control-Request-Headers,Access-Control-Request-Method,organisation-id'",
					"method.response.header.Access-Control-Allow-Origin":  "'*'",
				}},
		},
	}
	op.VendorExtensible.AddExtension("x-amazon-apigateway-integration", extension)
}

// Remove all the unwanted tags or param not supported by AWS API Gateway REST API
func applyFilters(swagger *swg.Swagger) {
	definitions := swagger.Definitions
	for key, schema := range swagger.Definitions {
		log.WithFields(log.Fields{
			"dto": key,
		}).Info("Filtering the `example` tag for key")
		removeTagExample(&schema.SchemaProps)
		removeNonAlphanumericModelsWithNoReferendes(key, definitions)
		renameNonAlphanumericModels(key, definitions)
	}
}

func renameNonAlphanumericModels(key string, definitions swg.Definitions) {
	if alphaNumRegexp.MatchString(key) {
		newKey := alphaNumRegexp.FindAllStringSubmatch(key, -1)[0][2]

		// check if the definition has non alphanumeric reference too
		definition := definitions[key]
		renameNonAlphanumeric(&definition)

		// reassign the schema to alphanumeric key
		definitions[newKey] = definitions[key]

		// delete the existing schema referende with non alphanumeric key
		delete(definitions, key)
	}
}

func renameNonAlphanumeric(schema *swg.Schema) {
	if schema.SchemaProps.Properties != nil {
		for _, sch := range schema.SchemaProps.Properties {
			if sch.Items != nil {
				ref := sch.Items.Schema.SchemaProps.Ref.Ref
				url := ref.GetURL()
				if alphaNumRegexp.MatchString(url.Fragment) {
					url.Fragment = strings.ReplaceAll(url.Fragment, "model.", "")
					fmt.Println(url.Fragment)
				}
			}

		}
	}
}

// Remove non alphanumeric models. These models are not been referencd in swagger definition and causes issues wih aws apigw
func removeNonAlphanumericModelsWithNoReferendes(key string, definitions swg.Definitions) {
	if strings.Contains(key, "»") {
		delete(definitions, key)
	}
}

// Filters the tag example
func removeTagExample(props *swg.SchemaProps) {
	for key, schema := range props.Properties {
		log.WithFields(log.Fields{"key": key}).Info("Removing example tag from ")
		schema.SwaggerSchemaProps.Example = nil
		props.Properties[key] = schema
	}
}

func isPathVisible(path swg.PathItem) bool {
	values := reflect.ValueOf(path.PathItemProps)
	num := values.NumField()
	for i := 0; i < num; i++ {
		if values.Field(i).IsNil() {
			continue
		}
		value := values.Field(i).Interface().(*swg.Operation)
		str, ok := value.Extensions.GetString("x-publish")

		if !ok {
			return true
		}
		getBool, err := strconv.ParseBool(str)
		if err == nil {
			return getBool
		}
	}
	return true
}

func isSecurityEnabled(path swg.PathItem) bool {
	values := reflect.ValueOf(path.PathItemProps)
	num := values.NumField()
	for i := 0; i < num; i++ {
		if values.Field(i).IsNil() {
			continue
		}
		value := values.Field(i).Interface().(*swg.Operation)
		str, ok := value.Extensions.GetString("x-auth-disabled")
		if !ok {
			return true
		}

		getBool, err := strconv.ParseBool(str)
		if err == nil {
			return getBool
		}
	}
	return true
}
