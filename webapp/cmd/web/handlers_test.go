package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var pageTests = []struct {
	name               string
	url                string
	expectedStatusCode int
}{
	{name: "home", url: "/", expectedStatusCode: http.StatusOK},
	{name: "404", url: "/forced404", expectedStatusCode: http.StatusNotFound},
}

func Test_application_home(t *testing.T) {

	routes := app.routes()

	//spawn test http server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	//range through the test data
	for _, e := range pageTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("%s: expected status %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}

}

var homeTests = []struct {
	name         string
	putInSession string
	expectedHTML string
}{
	{name: "first visit", putInSession: "", expectedHTML: "<small>From Session:"},
	{name: "second visit", putInSession: "hello, world!", expectedHTML: "<small>From Session: hello, world!"},
}

func TestAppHome(t *testing.T) {
	for _, e := range homeTests {
		req, _ := http.NewRequest("GET", "/", nil)
		req = addContextAndSessionToRequest(req, app)
		_ = app.Session.Destroy(req.Context())

		if e.putInSession != "" {
			app.Session.Put(req.Context(), "test", e.putInSession)
		}

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.Home)

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("%s: returned wrong status code expected 2-- but got %d", e.name, rr.Code)
		}

		body, _ := io.ReadAll(rr.Body)

		if !strings.Contains(string(body), e.expectedHTML) {
			t.Errorf("%s: did not find correct text in HTML, expected %s", e.name, e.expectedHTML)
		}
	}
}

func TestApp_renderWithBadTemplate(t *testing.T) {
	// set template path to a location with a bad template
	pathToTemplates = "./testdata/"

	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app)
	rr := httptest.NewRecorder()

	err := app.render(rr, req, "bad.page.gohtml", &TemplateData{})
	if err == nil {
		t.Error("expected error and got none")
	}
}

func getCtx(req *http.Request) context.Context {
	return context.WithValue(req.Context(), contextUserKey, "unknown")
}

func addContextAndSessionToRequest(req *http.Request, app application) *http.Request {
	req = req.WithContext(getCtx(req))

	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}
