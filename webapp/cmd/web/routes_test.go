package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

var registered = []struct {
	route  string
	method string
}{
	{route: "/", method: "GET"},
	{route: "/static/*", method: "GET"},
	{route: "/login", method: "POST"},
	{route: "/user/profile", method: "GET"},
}

func Test_application_routes(t *testing.T) {

	mux := app.routes()

	chiRoutes := mux.(chi.Routes)

	for _, route := range registered {
		if !routeExists(route.route, route.method, chiRoutes) {
			t.Errorf("route %s is not registered", route.route)
		}
	}
}

func routeExists(testRoute string, testMethod string, chiRoutes chi.Routes) bool {
	found := false
	_ = chi.Walk(chiRoutes, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if strings.EqualFold(method, testMethod) && strings.EqualFold(route, testRoute) {
			found = true
		}
		return nil
	})
	return found
}
