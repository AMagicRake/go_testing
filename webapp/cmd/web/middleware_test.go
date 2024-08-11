package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

var contextTests = []struct {
	headerName  string
	headerValue string
	addr        string
	emptyAddr   bool
}{
	{headerName: "", headerValue: "", addr: "", emptyAddr: false},
	{headerName: "", headerValue: "", addr: "", emptyAddr: true},
	{headerName: "X-Forwarded-For", headerValue: "192.2.2.1", addr: "", emptyAddr: false},
	{headerName: "", headerValue: "", addr: "hello:world", emptyAddr: false},
}

func Test_application_addIPToContext(t *testing.T) {

	// createa dummy handler that we'll use to checkthe context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//make sure that the value exists in the context
		val := r.Context().Value(contextUserKey)
		if val == nil {
			t.Error(contextUserKey, "not present")
		}

		// make sure we got a string back

		ip, ok := val.(string)
		if !ok {
			t.Error("not string")
		}
		t.Log(ip)

	})

	for _, e := range contextTests {
		handlerToTest := app.addIPToContext(nextHandler)

		req := httptest.NewRequest("GET", "http://testing", nil)

		if e.emptyAddr {
			req.RemoteAddr = ""
		}

		if len(e.headerName) > 0 {
			req.Header.Add(e.headerName, e.headerValue)
		}

		if len(e.addr) > 0 {
			req.RemoteAddr = e.addr
		}

		handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	}

}

func Test_application_ipFromContext(t *testing.T) {

	testValue := "1.1.1.1"
	ctx := context.WithValue(context.Background(), contextUserKey, testValue)

	res := app.ipFromContext(ctx)

	if res != testValue {
		t.Errorf("ipFromContext failed, expected %s, got %s", testValue, res)
	}
}

var authTests = []struct {
	name   string
	isAuth bool
}{
	{name: "logged in", isAuth: true},
	{name: "not logged in", isAuth: false},
}

func Test_application_auth(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	for _, e := range authTests {
		handlerToTest := app.auth(nextHandler)
		req := httptest.NewRequest("GET", "http://testing", nil)
		req = addContextAndSessionToRequest(req, app)
		if e.isAuth {
			app.Session.Put(req.Context(), "user", data.User{ID: 1})
		}
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if e.isAuth && rr.Code != http.StatusOK {
			t.Errorf("%s: expected status 200 but got %d", e.name, rr.Code)
		}

		if !e.isAuth && rr.Code != http.StatusSeeOther {
			t.Errorf("%s: expected 303 but got %d", e.name, rr.Code)
		}
	}

}
