package login

import (
	"log"
	"net/http"
	"os"
	"errors"

	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	GoogleOauth "google.golang.org/api/oauth2/v2"
	GooglePeople "google.golang.org/api/people/v1"
)

var(
	GoogleConfig = struct {
		ClientID string
		ClientSecret string
		RedirectURL string
		Endpoint oauth2.Endpoint
	}{
		ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL: os.Getenv("GOOGLE_REDIRECT_URL"),
		Endpoint: oauth2.Endpoint{
			"https://accounts.google.com/o/oauth2/v2/auth",
			"https://www.googleapis.com/oauth2/v4/token",
		},
	}
)

type Google struct {
	w	http.ResponseWriter
	r *http.Request
	config *oauth2.Config
	client *http.Client
}

func NewGoogle(w http.ResponseWriter, r *http.Request) *Google {
	config := &oauth2.Config{
		ClientID:    	GoogleConfig.ClientID,
		ClientSecret: GoogleConfig.ClientSecret,
		RedirectURL:  GoogleConfig.RedirectURL,
		Scopes:       []string{GoogleOauth.UserinfoEmailScope, GoogleOauth.UserinfoProfileScope},
		Endpoint:     GoogleConfig.Endpoint,
	}
	return &Google{w, r, config, nil}
}


func (this *Google) Login() {
	sess := NewSession(this.w, this.r)
	log.Printf("sess ID %s", sess.session.ID)
	// TODO Set state token For CSRF attack check
	url := this.config.AuthCodeURL(sess.session.ID, oauth2.ApprovalForce, oauth2.AccessTypeOnline)
	if url == "" {
		// err
	}
	log.Printf("url %s", url)
	http.Redirect(this.w, this.r, url, http.StatusFound)
}

func (this *Google) Callback(redirect string) {
	ctx := appengine.NewContext(this.r)
	code := this.r.FormValue("code")

	// TODO state check
	state := this.r.FormValue("state")
	log.Printf("code:%s state:%s", code, state)

	tok, err := this.config.Exchange(ctx, code)
	if err != nil {
		panic(err)
	}
	if tok.Valid() == false {
		panic(errors.New("vaild token"))
	}

	sess := NewSession(this.w, this.r)
	sess.Set("Google_AccessToken", tok.AccessToken)

	this.client = this.config.Client(ctx, tok)
	service, _ := GoogleOauth.New(this.client)
	tokenInfo, _ := service.Tokeninfo().AccessToken(tok.AccessToken).Context(ctx).Do()
	// if Decode idToken
	// idToken := tok.Extra("id_token").(string)

	// Google People API have to enable api
	// https://console.developers.google.com/apis/api/people.googleapis.com/overview
	p, err := this.getPeople()
	if err != nil {
		log.Printf("People Get me error:%#v", err)
		panic(err)
	}
	sess.Set("id", tokenInfo.UserId)
	sess.Set("email", tokenInfo.Email)
	sess.Set("name", p.Names[0].DisplayName)
	sess.Set("photo", p.Photos[0].Url)
	/*
	// Show paramaters
	for _, name := range p.Names {
		log.Printf("name: %#v", name)
	}
	for _, photo := range p.Photos {
		log.Printf("photo: %#v", photo)
	}
	*/
	sess.Save()
	http.Redirect(this.w, this.r, redirect, http.StatusFound)
}

func (this *Google) getPeople() (*GooglePeople.Person, error) {
	if this.client == nil {
		return nil, errors.New("client not set")
	}
	service, err := GooglePeople.New(this.client) // Service
	if err != nil {
		return nil, err
	}
	people, err := service.People.Get("people/me").PersonFields("names,photos").Do()
	if err != nil {
		return nil, err
	}
	log.Printf("People me %#v", people)
	return people, nil
}
