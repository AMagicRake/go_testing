package main

import (
	"net/http"
	"net/http/httptest"
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
	pathToTemplates = "./../../templates"
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
