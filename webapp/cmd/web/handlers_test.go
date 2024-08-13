package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"webapp/pkg/data"
)

var pageTests = []struct {
	name                    string
	url                     string
	expectedStatusCode      int
	expectedURL             string
	expectedFirstStatusCode int
}{
	{
		name:                    "home",
		url:                     "/",
		expectedStatusCode:      http.StatusOK,
		expectedURL:             "/",
		expectedFirstStatusCode: http.StatusOK,
	},
	{
		name:                    "404",
		url:                     "/forced404",
		expectedStatusCode:      http.StatusNotFound,
		expectedURL:             "/forced404",
		expectedFirstStatusCode: http.StatusNotFound,
	},
	{
		name:                    "profile",
		url:                     "/user/profile",
		expectedStatusCode:      http.StatusOK,
		expectedURL:             "/",
		expectedFirstStatusCode: http.StatusSeeOther,
	},
}

func Test_application_home(t *testing.T) {

	routes := app.routes()

	//spawn test http server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	//range through the test data
	for _, e := range pageTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("%s: expected status %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}

		if resp.Request.URL.Path != e.expectedURL {
			t.Errorf("%s: expected final url of %s, but got %s", e.name, e.expectedURL, resp.Request.URL.Path)
		}

		resp2, _ := client.Get(ts.URL + e.url)

		if resp2.StatusCode != e.expectedFirstStatusCode {
			t.Errorf("%s: expected first status of %d, but got %d", e.name, e.expectedFirstStatusCode, resp2.StatusCode)
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

	pathToTemplates = "./../../templates/"
}

func getCtx(req *http.Request) context.Context {
	return context.WithValue(req.Context(), contextUserKey, "unknown")
}

func addContextAndSessionToRequest(req *http.Request, app application) *http.Request {
	req = req.WithContext(getCtx(req))

	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}

var loginTests = []struct {
	name               string
	postedData         url.Values
	expectedStatusCode int
	expectedLoc        string
}{
	{
		name: "valid login",
		postedData: url.Values{
			"email":    {"admin@example.com"},
			"password": {"secret"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLoc:        "/user/profile",
	},
	{
		name: "missing form data",
		postedData: url.Values{
			"email":    {""},
			"password": {""},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLoc:        "/",
	},
	{
		name: "bad credentials",
		postedData: url.Values{
			"email":    {"admin_bad@example.com"},
			"password": {"password"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLoc:        "/",
	},
	{
		name: "invalid credentials",
		postedData: url.Values{
			"email":    {"admin@example.com"},
			"password": {"password"},
		},
		expectedStatusCode: http.StatusSeeOther,
		expectedLoc:        "/",
	},
}

func Test_app_login(t *testing.T) {
	for _, e := range loginTests {
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(e.postedData.Encode()))
		req = addContextAndSessionToRequest(req, app)
		req.Header.Set("content-type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.Login)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: returned wrong status code, expected %d but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		actualLoc, err := rr.Result().Location()
		if err == nil {
			if actualLoc.String() != e.expectedLoc {
				t.Errorf("%s: expected location %s, but got %s", e.name, e.expectedLoc, actualLoc.String())
			}
		} else {
			t.Errorf("%s: no location header set", e.name)
		}

	}
}

func Test_app_UploadFiles(t *testing.T) {
	// set up pipes
	pr, pw := io.Pipe()

	// create new writer for *io.Writer
	writer := multipart.NewWriter(pw)

	// create a wg
	wg := &sync.WaitGroup{}

	wg.Add(1)

	// run go routine to simulate uploading file
	go simulatePNGUpload("./testdata/img.png", writer, t, wg)

	// read from pipe
	req := httptest.NewRequest("POST", "/", pr)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	// call uploadfiles
	uploadedFiles, err := app.UploadFiles(req, "./testdata/uploads/")
	if err != nil {
		t.Error(err)
	}

	// perform tests
	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].OriginalFileName)); err != nil {
		t.Errorf("file failed to upload: %s", err.Error())
	}

	// cleanup
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].OriginalFileName))

	wg.Wait()
}

func simulatePNGUpload(fileToUpload string, writer *multipart.Writer, t *testing.T, wg *sync.WaitGroup) {
	defer writer.Close()
	defer wg.Done()

	// create form data field file with value being filename

	part, err := writer.CreateFormFile("file", path.Base(fileToUpload))
	if err != nil {
		t.Error(err)
	}

	// open actual file
	file, err := os.Open(fileToUpload)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	//decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		t.Error("error decoding image:", err)
	}

	//write the png to the pipe
	err = png.Encode(part, img)
	if err != nil {
		t.Error(err)
	}

}

func Test_app_UploadProfilePicture(t *testing.T) {
	pathToUploads = "./testdata/uploads"
	filePath := "./testdata/img.png"

	fieldName := "file"

	body := new(bytes.Buffer)

	mw := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}

	w, err := mw.CreateFormFile(fieldName, filePath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(w, file); err != nil {
		t.Fatal(err)
	}

	mw.Close()

	req := httptest.NewRequest("POST", "/", body)
	req = addContextAndSessionToRequest(req, app)

	app.Session.Put(req.Context(), "user", data.User{ID: 1})
	req.Header.Add("Content-Type", mw.FormDataContentType())

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(app.UploadProfilePicture)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code: %d", rr.Code)
	}

	_ = os.Remove("./testdata/uploads/img.png")
}
