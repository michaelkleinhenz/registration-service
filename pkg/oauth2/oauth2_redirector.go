package oauth2

import (
	"log"
	"net/http"
)

// Redirector is a redirect handler for incoming registration requests.
type Redirector struct {
	ClientID            string
	ClientSecret        string
	OIDAuthorizationURL string
	CallbackURL         string
}

func (h *Redirector) createRedirectURL() string {
	return h.OIDAuthorizationURL + "?response_type=code&client_id=" + h.ClientID + "&redirect_uri=" + h.CallbackURL + "&scope=read"
}

// ServeHTTP reroutes the client to the authorization URL
func (h Redirector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Redirecting client to RHD authorization url..")
	http.Redirect(w, r, h.createRedirectURL(), http.StatusSeeOther)
}
