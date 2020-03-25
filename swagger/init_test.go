package swagger

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	// set all env vars
	os.Setenv(AuthType, "apiKey")
	os.Setenv(AuthUrl, ExpectedAuthorizerArn)
	os.Setenv(ApiGwName, "api-gw-dev")
	os.Setenv(AuthName, ExpectedAuthorizerName)
	os.Setenv(CorsEnabled, "true")
	os.Setenv(CustomHeaders, "X-JWT-Assertion,organisation-id")

	// Run the test suite
	retCode := m.Run()

	// call with result of m.Run()
	os.Exit(retCode)
}
