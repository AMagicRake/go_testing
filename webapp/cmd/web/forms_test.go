package main

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func Test_form_Has(t *testing.T) {
	form := NewForm(nil)
	has := form.Has("whatever")
	if has {
		t.Error("form shows has field when it should not")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = NewForm(postedData)

	has = form.Has("a")
	if !has {
		t.Error("form missing field")
	}
}

func Test_form_required(t *testing.T) {
	req := httptest.NewRequest("POST", "https://test.com", nil)
	form := NewForm(req.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	req = httptest.NewRequest("POST", "https://test.com", nil)
	req.PostForm = postedData
	form = NewForm(req.PostForm)

	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("shows post does not have required fields when it does")
	}

}
