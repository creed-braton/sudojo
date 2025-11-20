package server

import (
	"fmt"
	"net/http"
	"sudojo/service/tenant"

	"github.com/gorilla/websocket"
)

type Server interface {
	Listen() error
}

type server struct {
	port     string
	origin   string
	router   *http.ServeMux
	upgrader websocket.Upgrader
	tenant   tenant.Service
}

func New(port, origin string, tenant tenant.Service) *server {
	s := &server{
		port:   port,
		origin: origin,
		router: http.NewServeMux(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		tenant: tenant,
	}

	routes := map[string]map[string]http.HandlerFunc{
		"/lobbies": {
			"POST": s.postLobby,
		},
		"/lobbies/{id}": {
			"GET":   s.getLobby,
			"PATCH": s.patchLobby,
		},
	}

	for path, route := range routes {
		handler := s.methodRouter(route)
		if len(origin) > 0 {
			handler = s.cors(handler, route)
		}
		s.router.Handle(fmt.Sprintf("/api%s", path), handler)
	}
	return s
}

func (s *server) Listen() error {
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), s.router)
}

func (s *server) methodRouter(methods map[string]http.HandlerFunc) http.Handler {
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

func (s *server) cors(
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
