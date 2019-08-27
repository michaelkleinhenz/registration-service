package oauth2_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/codeready-toolchain/registration-service/pkg/configuration"
	"github.com/codeready-toolchain/registration-service/pkg/oauth2"
	"github.com/stretchr/testify/assert"
)

func TestOAuth2Handler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/api/callback", nil)
	if err != nil {
		t.Fatal(err)
	}

	// create logger and registry.
	logger := log.New(os.Stderr, "", 0)
	configRegistry := configuration.CreateEmptyRegistry()

	// set the config for testing mode, the handler may use this.
	configRegistry.GetViperInstance().Set("testingmode", true)
	assert.True(t, configRegistry.IsTestingMode(), "testing mode not set correctly to true")
	configRegistry.GetViperInstance().Set("version", "0.0.0-testingmode")

	// create handler instance.
	oauth2Service := oauth2.New(logger, configRegistry)
	handler := http.HandlerFunc(oauth2Service.OAuth2Handler)

	t.Run("not implemented", func(t *testing.T) {	
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)
		// Check the status code is what we expect.
		assert.Equal(t, rr.Code, http.StatusInternalServerError, "handler returned wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
	})

}
