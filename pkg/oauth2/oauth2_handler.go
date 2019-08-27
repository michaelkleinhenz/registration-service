package oauth2

import (
	"log"
	"net/http"

	"github.com/codeready-toolchain/registration-service/pkg/configuration"
)

// Service implements the service health endpoint.
type Service struct {
	config *configuration.Registry
	logger *log.Logger
}

// New returns a new healthService instance.
func New(logger *log.Logger, config *configuration.Registry) *Service {
	r := new(Service)
	r.logger = logger
	r.config = config
	return r
}

// OAuth2Handler handles OAuth2 redirect responses.
func (srv *Service) OAuth2Handler(w http.ResponseWriter, r *http.Request) {
	/*
	w.Header().Set("Content-Type", "application/json")
	healthInfo := srv.getHealthInfo()
	if healthInfo["alive"].(bool) {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	err := json.NewEncoder(w).Encode(healthInfo)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	*/
	http.Error(w, "not implemented yet", http.StatusInternalServerError)
	return
}
