package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"webapp/pkg/data"

	"github.com/go-chi/chi/v5"
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

func TestApi_refresh(t *testing.T) {
	tests := []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRefreshTime   bool
	}{
		{"valid", "", http.StatusOK, true},
		{"too early", "", http.StatusTooEarly, false},
		{"expired token", expiredToken, http.StatusBadRequest, false},
	}

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, e := range tests {
		var tkn string
		if e.token == "" {
			if e.resetRefreshTime {
				refreshTokenExpiry = time.Second * 1
			}
			tokens, _ := app.generateTokenPair(&testUser)
			tkn = tokens.RefreshToken
		} else {
			tkn = e.token
		}

		postedData := url.Values{
			"refresh_token": {tkn},
		}

		req := httptest.NewRequest("POST", "/refresh-token", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.refresh)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		refreshTokenExpiry = oldRefreshTime
	}
}

func TestApi_userEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		json           string
		paramID        string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{"all users", "GET", "", "", app.allUsers, http.StatusOK},
		{"delete user", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"delete user bad url param", "DELETE", "", "XD", app.deleteUser, http.StatusBadRequest},
		{"get user valid", "GET", "", "1", app.getUser, http.StatusOK},
		{"get user invalid", "GET", "", "2", app.getUser, http.StatusBadRequest},
		{"get user invalid url param", "GET", "", "F", app.getUser, http.StatusBadRequest},
		{
			"update valid user",
			"PATCH",
			`{"id":1, "first_name": "Administrator", "last_name":"User", "email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusNoContent,
		},
		{
			"update invalid user",
			"PATCH",
			`{"id":2, "first_name": "Administrator", "last_name":"User", "email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"update invalid json",
			"PATCH",
			`{"id":1, first_name: "Administrator", "last_name":"User", "email":"admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},
		{
			"insert valid user",
			"PUT",
			`{"first_name": "Jack", "last_name":"Smith", "email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusNoContent,
		},
		{
			"insert invalid user",
			"PUT",
			`{"foo":"bar", "first_name": "Jack", "last_name":"Smith", "email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
		{
			"insert invalid json user",
			"PUT",
			`{"foo":"bar""first_name": "Jack", "last_name":"Smith", "email":"jack@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
	}

	for _, e := range tests {
		var req *http.Request
		if e.json == "" {
			req = httptest.NewRequest(e.method, "/", nil)
		} else {
			req = httptest.NewRequest(e.method, "/", strings.NewReader(e.json))
		}

		if e.paramID != "" {
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", e.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(e.handler)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatus {
			t.Errorf("%s: wrong status returned, expected %d but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}
