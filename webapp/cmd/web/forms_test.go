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

func Test_form_check(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is required")

	if form.Valid() {
		t.Error("valid returns true and it should be false when calling check")
	}

}

func Test_form_ErrorGet(t *testing.T) {
	form := NewForm(nil)
	form.Check(false, "password", "password is required")

	s := form.Errors.Get("password")
	if len(s) == 0 {
		t.Error("should have an error and don't")
	}

	s = form.Errors.Get("userName")
	if len(s) != 0 {
		t.Error("have error when we shouldnt")
	}
}
