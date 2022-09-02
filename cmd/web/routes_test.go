package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"testing"
)

var routes = []string{
	"/",
	"/login",
	"/login",
	"/logout",
	"/register",
	"/register",
	"/activate",
	"/plans",
	"/subscribe",
}

func Test_Routes_Exist(t *testing.T) {
	testRoutes := testApp.routes()

	chiRoutes := testRoutes.(chi.Router)
	for _, route := range routes {
		routeExist(t, chiRoutes, route)
	}
}

func routeExist(t *testing.T, routes chi.Router, route string) {
	found := false

	_ = chi.Walk(routes, func(method string, foundRoute string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == foundRoute {
			found = true
		}
		return nil
	})

	if !found {
		t.Errorf("Did not find %s in registered routes", route)
	}
}
