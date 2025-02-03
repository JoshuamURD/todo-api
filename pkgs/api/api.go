package api

import (
	"context"
	"net/http"
)

type EndpointMap map[string]func(ctx context.Context, req *http.Request) (interface{}, error)

type API struct {
	endpoints EndpointMap
	server    *http.Server
	router    *http.ServeMux
	port      string
	host      string
}

func NewAPI(port, host string) *API {
	return &API{
		endpoints: make(EndpointMap),
		server:    &http.Server{},
		router:    http.NewServeMux(),
		port:      port,
		host:      host,
	}
}

func (a *API) RegisterRoute(endpoint string, func http.HandlerFunc) error {
	return nil
}
