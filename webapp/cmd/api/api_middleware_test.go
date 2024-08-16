package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func TestMiddleware_enableCORS(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	tests := []struct {
		name         string
		method       string
		expectHeader bool
	}{
		{"preflight", "OPTIONS", true},
		{"get", "GET", false},
	}

	for _, e := range tests {
		handlerToTest := app.enableCORS(nextHandler)

		req := httptest.NewRequest(e.method, "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if e.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") == "" {
			t.Errorf("%s: expected a header but did not get ones", e.name)
		}

		if !e.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Errorf("%s: did not expect a header, but got one", e.name)
		}

	}
}

func TestMiddleware_authRequired(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	var tests = []struct {
		name               string
		token              string
		expectedAuthorized bool
		setHeader          bool
	}{
		{"valid token", fmt.Sprintf("Bearer %s", tokens.Token), true, true},
		{"no token", "", false, false},
		{"invalid token", fmt.Sprintf("Bearer %s", expiredToken), false, true},
	}

	for _, e := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		if e.setHeader {
			req.Header.Set("Authorization", e.token)
		}

		rr := httptest.NewRecorder()

		handlerToTest := app.authRequired(nextHandler)
		handlerToTest.ServeHTTP(rr, req)

		if e.expectedAuthorized && rr.Code == http.StatusUnauthorized {
			t.Errorf("%s: got code 401 and should not have", e.name)
		}

		if !e.expectedAuthorized && rr.Code != http.StatusUnauthorized {
			t.Errorf("%s: user authorized and should not have been", e.name)
		}

	}

}
