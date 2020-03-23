# Swagger Publisher

This tool publishes REST API into API Gateway. This tool is packed as a docker container and can be run in a `ci pipeline` after a given service
is deployed to the target environment. This tool fetches the swagger document from a given `url`, render it by adding all the required AWS extensions
and publish it to the API gateway 

When the tool is run, it carries out the following tasks:

* Fetches vanilla swagger document from a given deployed service - exp: http://internal-api.dev.astoapp.co.uk/account-service/v2/api-docs
* Renders the swagger document so as the `example` tag is removed, allow all mimetype (*/*) to application/json
* Add aws extensions - see https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-integration.html to the swagger doc
* Import the rendered swagger into API Gateway - this will create the api gateway resources.
* Deploy all the resources created above into the given `stage`

## API Extensions

In order to control this tool on deployment, there are a few Swagger Extensions we leverage as configuration.

We have aimed to go for secure-by-default.

* `x-asto-publish` - this flag if set to false, the endpoint will not get published.
* `x-asto-auth-disabled` - 

In Java these extensions can be controlled using something similar to the below, simply add this annotation above a controller method:

```
@ApiOperation(value = "Describe this API for the Swagger docs",
        extensions = @Extension(properties = {
            @ExtensionProperty(name = "x-asto-publish", value = "false"),
            @ExtensionProperty(name = "x-asto-auth-disabled", value = "false"),
        }))
```

## Required environment variables

These are the environment variables required for this tool. By default, the tool uses the environment variables set for the `dev` environment
* `VPC_LINK_ID`: The vpc link Id for a given environment
* `CONNECTION_TYPE`: the connection type: `VPC_LINK` or `PUBLIC`(http)
* `API_GATEWAY_ID`: the api gateway Id
* `STAGE_NAME`: the stage name of the api gateway on which the resources will get deployed to
* `API_GATEWAY_NAME`: The api gateway name
* `AUTH_URL`: If `custom` authentication is enabled on the endpoints then the `authentcation url` is required `- more details in the auth section below`
* `AUTH_NAME`: The authorizer name
* `AUTH_TYPE`: Currently only two types are supported: `apiKey` and `oauth2(jwt token - cognito)`
* `SWAGGER_URL`: The url of the swagger documentation
* `AWS_ACCESS_KEY_ID`: the aws access key
* `AWS_SECRET_ACCESS_KEY`: the aws secret access key
* `ASSUME_ROLE`: the aws assume role that allow rest api into the api gateway
* `ENDPOINT_URL`: the endpoint url of the service. This should be listed in the json document as host+basePath and used by the publisher if this env var is not provided. 

## Running the publisher

1. Running the publisher with `public (http)` connection type

docker run --env API_GATEWAY_NAME=api-gw-dev \
--env CONNECTION_TYPE=VPC_LINK \
--env AUTH_URL=arn:aws:apigateway:eu-west-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-1:762587740553:function:api-gateway-authorizer-dev-auth/invocations \
--env AUTH_NAME=api-gw-dev \
--env AUTH_TYPE=apiKey \
--env API_GATEWAY_ID=nildadqvqg \
--env STAGE_NAME=v1 \
--env ASSUME_ROLE=arn:aws:iam::74***87740553:role/apigw-role \
--env AWS_ACCESS_KEY_ID=AKIA3DDOFPG****CYL \
--env AWS_SECRET_ACCESS_KEY=8uT/29******* apigw/publisher /bin/aws-apigw-publisher
 
2. Running the publisher `vpc link` connection type

docker run --env API_GATEWAY_NAME=wave-api-gw-dev \
--env CONNECTION_TYPE=VPC_LINK \
--env  VPC_LINK_ID=596jx1 \
--env AUTH_URL=arn:aws:apigateway:eu-west-1:lambda:path/2015-03-31/functions/arn:aws:lambda:eu-west-1:762587740553:function:api-gateway-authorizer-dev-auth/invocations \
--env AUTH_NAME=wave-api-gw-dev \
--env AUTH_TYPE=apiKey \
--env API_GATEWAY_ID=nildadqvqg \
--env STAGE_NAME=v1 \
--env API_GATEWAY_ID=nildadqvqg \
--env STAGE_NAME=v1 \
--env ASSUME_ROLE=arn:aws:iam::74***87740553:role/apigw-role \
--env AWS_ACCESS_KEY_ID=AKIA3DDOFPG****CYL \
--env AWS_SECRET_ACCESS_KEY=8uT/29******* apigw/publisher /bin/aws-apigw-publisher



## Authorization schemes

Details on the authorization scheme can be found here: https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-swagger-extensions-authorizer.html

