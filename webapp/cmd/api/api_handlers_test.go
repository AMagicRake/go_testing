package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var authTests = []struct {
	name               string
	requestBody        string
	expectedStatusCode int
}{
	{"valid user", `{"email":"admin@example.com","password":"secret"}`, http.StatusOK},
	{"not json", `i'm not json`, http.StatusUnauthorized},
	{"empty json", `{}`, http.StatusUnauthorized},
	{"empty email", `{"email":""}`, http.StatusUnauthorized},
	{"empty password", `{"email":"admin@example.com"}`, http.StatusUnauthorized},
	{"invalid user", `{"email":"admin@nothere.com","password":"secret"}`, http.StatusUnauthorized},
}

func TestApi_authenticate(t *testing.T) {

	for _, e := range authTests {
		var reader io.Reader
		reader = strings.NewReader(e.requestBody)

		req, _ := http.NewRequest("POST", "/", reader)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.authenticate)

		handler.ServeHTTP(rr, req)

		if e.expectedStatusCode != rr.Code {
			t.Errorf("%s: returned wrong status code expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

	}
}
