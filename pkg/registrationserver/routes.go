package registrationserver

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/codeready-toolchain/registration-service/pkg/health"
	"github.com/codeready-toolchain/registration-service/pkg/oauth2"
	"github.com/codeready-toolchain/registration-service/pkg/static"
)

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	Assets http.FileSystem
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// no absolute path, respond with a 400 bad request and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if the file exists in the assets
	_, err = h.Assets.Open(path)
	if err != nil {
		// file does not exist, redirect to index
		log.Printf("File %s does not exist.", path)
		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(h.Assets).ServeHTTP(w, r)
}

// SetupRoutes registers handlers for various URL paths. You can call this
// function more than once but only the first call will have an effect.
func (srv *RegistrationServer) SetupRoutes() error {
	var err error
	srv.routesSetup.Do(func() {

		// /status is something you should always have in any of your services,
		// please leave it as is.
		healthService := health.New(srv.logger, srv.Config())

		// heath service
		srv.router.HandleFunc("/api/health", healthService.HealthCheckHandler).
			Name("health").
			Methods("GET")

		// callback service for oid2 flows
		srv.router.HandleFunc("/api/callback", healthService.HealthCheckHandler).
			Name("callback").
			Methods("GET")

		// ADD YOUR OWN ROUTES HERE

		// create the route for the oauth redirect
		oauth2 := oauth2.Redirector {
			ClientID:            srv.Config().GetOIDClientID(),
			ClientSecret:        srv.Config().GetOIDClientSecret(),
			OIDAuthorizationURL: srv.Config().GetOIDAuthorizationURL(),
			// TODO: this needs to be the registration app address!
			CallbackURL: "https://this.server.address/api/callback",
		}
		srv.router.PathPrefix("/redirect").Handler(oauth2)

		// create the route for static content, served from /
		spa := spaHandler{Assets: static.Assets}
		srv.router.PathPrefix("/").Handler(spa)
	})
	return err
}
