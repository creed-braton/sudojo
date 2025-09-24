package server

import (
	"net/http"
	"sudojo/service"
)

type Server struct {
	port     string
	router   *http.ServeMux
	services []service.Service
}

func New(port string, services []service.Service) *Server {
	server := &Server{
		port:   port,
		router: http.NewServeMux(),
	}

	for _, s := range services {
		for path, route := range s.Routes() {
			server.router.Handle(path, server.methodRouter(route))
		}
	}

	return server
}

func (s *Server) Listen() error {
	return http.ListenAndServe(s.port, s.router)
}

func (s *Server) methodRouter(methods map[string]http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if methods[r.Method] == nil {
			http.Error(w, "method not allowed", 405)
			return
		}
		methods[r.Method](w, r)
	})
}
