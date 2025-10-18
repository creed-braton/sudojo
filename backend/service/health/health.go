package health

import "net/http"

type service struct{}

func New() *service {
	return &service{}
}

func (s *service) Routes() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/health": {
			"GET": s.getHealth,
		},
	}
}

func (s *service) getHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("healthy"))
}
