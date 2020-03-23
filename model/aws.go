package model

// AWSAPIGatewayIntegration the custom x-amazon-apigateway-integration added to our Swagger Docs to tell API GW what to do
type AWSAPIGatewayIntegration struct {
	URI                 string                            `json:"uri"`
	ConnectionType      string                            `json:"connectionType"`
	ConnectionID        string                            `json:"connectionId"`
	IntegrationType     string                            `json:"type"`
	HTTPMethod          string                            `json:"httpMethod"`
	PassthroughBehavior string                            `json:"passthroughBehavior"`
	RequestParameters   map[string]string                 `json:"requestParameters"`
	RequestTemplates    map[string]string                 `json:"requestTemplates"`
	Responses           map[string]map[string]interface{} `json:"responses"`
}

// AWSSecurityDefinition the custom security definition used by API GW
