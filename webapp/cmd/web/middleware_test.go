package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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
