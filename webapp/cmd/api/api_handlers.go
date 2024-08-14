package main

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	UserName string `json:"email"`
	Password string `json:"password"`
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	creds := Credentials{}
	//read a json payload
	err := app.readJSON(w, r, &creds)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized 1"), http.StatusUnauthorized)
		return
	}

	//look up the user by email address
	user, err := app.DB.GetUserByEmail(creds.UserName)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized 2"), http.StatusUnauthorized)
		return
	}

	//check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized 3"), http.StatusUnauthorized)
		return
	}

	//generate tokens
	tokenPairs, err := app.generateTokenPair(user)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized 4"), http.StatusUnauthorized)
		return
	}

	//send tokens to user
	_ = app.writeJSON(w, http.StatusOK, tokenPairs)
}

func (app *application) refresh(w http.ResponseWriter, r *http.Request) {

}

func (app *application) allUsers(w http.ResponseWriter, r *http.Request) {

}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) insertUser(w http.ResponseWriter, r *http.Request) {

}
