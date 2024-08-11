package main

import (
	"os"
	"testing"
)

var app application

func TestMain(m *testing.M) {

	app.Session = getSession()
	pathToTemplates = "./../../templates/"

	os.Exit(m.Run())
}
