
# AWS Swagger Publisher

![Master CI](https://github.com/akhettar/aws-apigw-publisher/workflows/Master%20CI/badge.svg)
![](https://img.shields.io/docker/pulls/ayache/apigw-publisher)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/4839ba123ed9492aad75e84b5d89f4cf)](https://app.codacy.com/manual/akhettar/aws-apigw-publisher?utm_source=github.com&utm_medium=referral&utm_content=akhettar/aws-apigw-publisher&utm_campaign=Badge_Grade_Dashboard)

![Work in progress](pushing-cart.png)

This tool publishes REST API to AWS API Gateway from a given swagger document. It is packed as a docker container so it can be run in a `continuious integration pipeline`. This tool fetches the swagger document from a given `url`, render it by adding all the required AWS extensions
and publish it to AWS API gateway.

When the tool is run, it carries out the following tasks:

* Fetches vanilla swagger document from a given deployed service - exp: `https://raw.githubusercontent.com/swagger-api/swagger-spec/master/examples/v2.0/json/petstore-expanded.json`
* Add aws extensions - see [AWS extensions](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-integration.html) to the swagger doc
* Import the rendered swagger into API Gateway - this will create the api gateway resources.
* Deploy all the resources created above into the given `stage`

## API Extensions

In order to control this tool on deployment, there are a few Swagger Extensions we leverage as configuration.

* `x-publish` - this flag if set to false, the endpoint will not get published.
* `x-auth-disabled` - this flag if set to true, the endpoint will not be secured if custom auth is required

In Java these extensions can be controlled using something similar to the below, simply add this annotation above a controller method:

```
@ApiOperation(value = "Describe this API for the Swagger docs",
        extensions = @Extension(properties = {
            @ExtensionProperty(name = "x-publish", value = "false"),
            @ExtensionProperty(name = "x-auth-disabled", value = "true"),
        }))
```

## Required environment variables

These are the environment variables required for this tool. 

| Env variable              | Description           | Required  |
| -------------             |-------------          | ---------|
| `API_GATEWAY_ID`          | The api gateway Id    | Yes       |
| `API_GATEWAY_NAME`        | The api gateway name  | Yes       |
|`CONNECTION_TYPE`          | The integration type, the following connection type supported: `VPC_LINK`, `PUBLIC`(HTTP)| No (`PUBLIC`) is used by default)|
| `VPC_LINK_ID`             | The vpc link Id for a given environment    | No, required only if the connection type is of VPC link type       |
| `STAGE_NAME`              | The api gateway stage name for the resource to be deployed to    | Yes       |
| `AUTH_URL`                | If `custom` authentication is enabled on the endpoints then the `authentcation url` is required `- more details in the auth section below`    | No       |
| `AUTH_NAME`               | The authorizer name, see below the endpoint auth section for more details   | No       |
| `AUTH_TYPE`               | Currently only the custom auth is supported `apiKey`    | No       |
| `SWAGGER_URL`             | The url of the swagger document that can be sourced from     | Yes       |
| `AWS_ACCESS_KEY_ID`       | The aws access key    | Yes       |
| `AWS_SECRET_ACCESS_KEY`   | The aws secret access key    | Yes       |
| `ASSUME_ROLE`             | The assume role in arn format that allow this tool to publish the rest endpoints to api gateway   | Yes       |
| `ENDPOINT_URL`            | The internal host and the base endpoint of the service exp :`petstore.swagger.io/api`             | Yes       |
| `CORS_ENABLED`            | If this flag is present, `cors` is enabled across all the endpoints    | No       |
| `API_GATEWAY_ID`          | The api gateway Id    | Yes       |
| `CUSTOM_HEADERS`          | A list of comma separated headers to be mapped in the http headers of the endpoint, exp: `CUSTOM_HEADERS=header1,header2`  | No       |


## Running the publisher

1. Running the publisher with `public (http)` connection type

```shell script
docker run --env API_GATEWAY_NAME=api-gw-dev \
--env SWAGGER_URL=https://raw.githubusercontent.com/swagger-api/swagger-spec/master/examples/v2.0/json/petstore-expanded.json \
--env ENDPOINT_URL=petstore.swagger.io/api \
--env CONNECTION_TYPE=VPC_LINK \
--env AUTH_URL=arn:aws:apigateway:eu-west-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-1:<aws-account-id>:function:api-gateway-authorizer-dev-auth/invocations \
--env AUTH_NAME=api-gw-authorizer-dev \
--env AUTH_TYPE=apiKey \
--env API_GATEWAY_ID=nizzzddqg \
--env API_GATEWAY_NAME=app-gateway-name \
--env STAGE_NAME=v1 \
--env ASSUME_ROLE=arn:aws:iam::****************:role/apigw-role \
--env AWS_ACCESS_KEY_ID=************ \
--env AWS_SECRET_ACCESS_KEY=******************** ayache/apigw-publisher /bin/aws-apigw-publisher
```
 
2. Running the publisher `vpc link` connection type

```shell script
docker run --env API_GATEWAY_NAME=wave-api-gw-dev \
--env SWAGGER_URL=https://raw.githubusercontent.com/swagger-api/swagger-spec/master/examples/v2.0/json/petstore-expanded.json \
--env ENDPOINT_URL=petstore.swagger.io/api \
--env CONNECTION_TYPE=VPC_LINK \
--env VPC_LINK_ID=226jx1 \
--env AUTH_URL=arn:aws:apigateway:eu-west-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-1:<aws-account-id>:function:api-gateway-authorizer-dev-auth/invocations \
--env AUTH_NAME=wave-api-gw-dev \
--env AUTH_TYPE=apiKey \
--env STAGE_NAME=v1 \
--env API_GATEWAY_ID=nilbbdqvqg \
--env STAGE_NAME=v1 \
--env ASSUME_ROLE=arn:aws:iam::74***87740553:role/apigw-role \
--env AWS_ACCESS_KEY_*************CYL \
--env AWS_SECRET_ACCESS_KEY=************************** ayache/apigw-publisher /bin/aws-apigw-publisher
```

## Running the publisher in a circleci pipeline

The step below should be run after a service is successfully deployed to a target environment so the REST API can be deployed from the latest swagger document

```yaml
defaults: &swaggerpublisher
  <<: *dir
  docker:
    - image: ayache/swagger-publisher

publish-apigw-dev:
    <<: *swaggerpublisher
    steps:
      - checkout
      - run:
          name: Publish swagger to api gateway
          command: |
            SWAGGER_URL=https://raw.githubusercontent.com/swagger-api/swagger-spec/master/examples/v2.0/json/petstore-expanded.json \
            ENDPOINT_URL=petstore.swagger.io/api \
            CONNECTION_TYPE=VPC_LINK \
            VPC_LINK_ID=226jx1 \
            AUTH_URL=arn:aws:apigateway:eu-west-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-1:<aws-account-id>:function:api-gateway-authorizer-dev-auth/invocations \
            AUTH_NAME=wave-api-gw-dev \
            AUTH_TYPE=apiKey \
            STAGE_NAME=v1 \
            API_GATEWAY_ID=nilbbdqvqg \
            STAGE_NAME=v1 \
            ASSUME_ROLE=arn:aws:iam::74***87740553:role/apigw-role \
            AWS_ACCESS_KEY_*************CYL \
            AWS_SECRET_ACCESS_KEY=************************** ayache/apigw-publisher /bin/aws-apigw-publisher /bin/swagger-publisher
```

## AWS IAM
This tool recommends that the AWS IAM user is created with virtually no permissions at all. The only permission given to this user is the ability to assume a role by which the IAM user is permitted to publish REST endpoints to API Gateway. Some details can be
found in [AWS IAM policy do](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-iam-policy-examples.html)

## Authorization schemes

This publisher is using the GO openAPI spec version 2.0 to parse the swagger and adds all the AWS extensions. The only authentication type supported by this tool is the custom auth - see below an illustration

```json
"securityDefinitions" : {
    "test-authorizer" : {
      "type" : "apiKey",                         // Required and the value must be "apiKey" for an API Gateway API.
      "name" : "Authorization",                  // The name of the header containing the authorization token.
      "in" : "header",                           // Required and the value must be "header" for an API Gateway API.
      "x-amazon-apigateway-authtype" : "oauth2", // Specifies the authorization mechanism for the client.
      "x-amazon-apigateway-authorizer" : {       // An API Gateway Lambda authorizer definition
        "type" : "token",                        // Required property and the value must "token"
        "authorizerUri" : "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:account-id:function:function-name/invocations",
        "authorizerCredentials" : "arn:aws:iam::account-id:role",
        "identityValidationExpression" : "^x-[a-z]+",
        "authorizerResultTtlInSeconds" : 60
      }
    }
  }
```

Unfortunately the support for `JWT Cognito authentication` is not supported in GO openApi spec 2.0, the 3.0 version has not been released yet see [open spec 3.0 issue](https://github.com/go-openapi/spec/issues/21). If you are securing the endpoint with 
`JWT Cognito` then unfortunately you will have to run `terraform` command to secure these endpoints - see an example below

The snippet below is taken from [terraform doc](https://www.terraform.io/docs/providers/aws/r/api_gateway_method.html)
```json
variable "cognito_user_pool_name" {}

data "aws_cognito_user_pools" "this" {
  name = "${var.cognito_user_pool_name}"
}

resource "aws_api_gateway_rest_api" "this" {
  name = "with-authorizer"
}

resource "aws_api_gateway_resource" "this" {
  rest_api_id = "${aws_api_gateway_rest_api.this.id}"
  parent_id   = "${aws_api_gateway_rest_api.this.root_resource_id}"
  path_part   = "{proxy+}"
}

resource "aws_api_gateway_authorizer" "this" {
  name          = "CognitoUserPoolAuthorizer"
  type          = "COGNITO_USER_POOLS"
  rest_api_id   = "${aws_api_gateway_rest_api.this.id}"
  provider_arns = ["${data.aws_cognito_user_pools.this.arns}"]
}

resource "aws_api_gateway_method" "any" {
  rest_api_id   = "${aws_api_gateway_rest_api.this.id}"
  resource_id   = "${aws_api_gateway_resource.this.id}"
  http_method   = "ANY"
  authorization = "COGNITO_USER_POOLS"
  authorizer_id = "${aws_api_gateway_authorizer.this.id}"

  request_parameters = {
    "method.request.path.proxy" = true
  }
}
```

Details on the authorization scheme can be found here: [AWS swagger extension authorizer](https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-authorizer.html)

