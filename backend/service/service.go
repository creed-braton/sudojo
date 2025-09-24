package service

import "net/http"

type Service interface {
	Routes() map[string]map[string]http.HandlerFunc
}
