package server

import (
	"fmt"
	"net/http"
	"sudojo/service"
)

type Server struct {
	port     string
	origin   string
	router   *http.ServeMux
	services []service.Service
}

func New(port, origin string, services []service.Service) *Server {
	server := &Server{
		port:   port,
		origin: origin,
		router: http.NewServeMux(),
	}

	for _, s := range services {
		for path, route := range s.Routes() {
			handler := server.methodRouter(route)
			if len(origin) > 0 {
				handler = server.cors(handler, route)
			}
			server.router.Handle(fmt.Sprintf("/api%s", path), handler)
		}
	}

	return server
}

func (s *Server) Listen() error {
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), s.router)
}

func (s *Server) methodRouter(methods map[string]http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			allow := "OPTIONS, "
			for m := range methods {
				allow += m + ", "
			}
			w.Header().Set("Allow", allow[:len(allow)-2])
			w.Header().Set("Cache-Control", "public, max-age=86400")
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte("no content"))
			return
		}

		if methods[r.Method] == nil {
			http.Error(w, "method not allowed", 405)
			return
		}

		methods[r.Method](w, r)
	})
}

func (s *Server) cors(
	next http.Handler,
	methods map[string]http.HandlerFunc,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Origin", s.origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			allow := "OPTIONS, "
			for m := range methods {
				allow += m + ", "
			}
			w.Header().Set("Access-Control-Allow-Methods", allow[:len(allow)-2])
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte("no content"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
