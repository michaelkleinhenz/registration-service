package oauth2_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codeready-toolchain/registration-service/pkg/oauth2"
	"github.com/stretchr/testify/assert"
)

func TestOAuth2Redirector(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/redirect", nil)
	if err != nil {
		t.Fatal(err)
	}

	// create handler instance.
	oauth2Redirector := oauth2.Redirector {
		ClientID:            "test-clientid",
		ClientSecret:        "test-clientsecret",
		OIDAuthorizationURL: "http://test-auth-url/",
		CallbackURL: 				 "http://test-callback-url/",
	}
	handler := http.HandlerFunc(oauth2Redirector.ServeHTTP)

	t.Run("not implemented", func(t *testing.T) {	
		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()
		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		handler.ServeHTTP(rr, req)
		// Check the status code is what we expect.
		assert.Equal(t, rr.Code, http.StatusSeeOther, "handler returned wrong status code: got %v want %v", rr.Code, http.StatusSeeOther)
		// TODO: check if the redirect URL contents are correct!
	})

}
