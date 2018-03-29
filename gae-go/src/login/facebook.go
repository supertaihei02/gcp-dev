package login

import (
	"log"
	"net/http"
	"os"
	"errors"
	"io"
	"strings"
	"encoding/base32"
	"crypto/rand"

	"golang.org/x/oauth2"
	"google.golang.org/appengine"
	"encoding/json"
	"net/url"
)

const(
	FacebookAccessTokenKey = "FacebookAccessToken"
	FacebookOauthState = "FacebookOauthState"
	FacebookApiEndpoint = "https://graph.facebook.com/v2.12/"
)

var(
	FacebookConfig = struct {
		ClientID string
		ClientSecret string
		RedirectURL string
		Endpoint oauth2.Endpoint
	}{
		ClientID: os.Getenv("FACEBOOK_APP_ID"),
		ClientSecret: os.Getenv("FACEBOOK_APP_SECRET"),
		RedirectURL: os.Getenv("FACEBOOK_REDIRECT_URL"),
		Endpoint: oauth2.Endpoint{
			"https://www.facebook.com/dialog/oauth",
			"https://graph.facebook.com/oauth/access_token",
		},
	}
)

type Facebook struct {
	w	http.ResponseWriter
	r *http.Request
	config *oauth2.Config
	tok *oauth2.Token
	client *http.Client
}

func NewFacebook(w http.ResponseWriter, r *http.Request) *Facebook {
	config := &oauth2.Config{
		ClientID:    	FacebookConfig.ClientID,
		ClientSecret: FacebookConfig.ClientSecret,
		RedirectURL:  FacebookConfig.RedirectURL,
		Scopes:       []string{"email", "user_about_me", "user_birthday"},
		Endpoint:     FacebookConfig.Endpoint,
	}

	fb := &Facebook{w, r, config, nil, nil}
	/*
	// TODO create token form session
	sess := NewSession(w, r, &sessionConfig)
	*/
	return fb
}


func (this *Facebook) Login() {
	sess := NewSession(this.w, this.r, &sessionConfig)
	log.Printf("sess ID %s", sess.session.ID)
	// CSRF attack check state
	b := make([]byte, 48)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	state := strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
	sess.Set(FacebookOauthState, state)
	sess.Save()
	log.Printf("before state:%s", state)

	url := this.config.AuthCodeURL(state, oauth2.ApprovalForce, oauth2.AccessTypeOnline)
	if url == "" {
		// err
	}
	log.Printf("url %s", url)
	http.Redirect(this.w, this.r, url, http.StatusFound)
}

func (this *Facebook) Callback(redirect string) {
	sess := NewSession(this.w, this.r, &sessionConfig)

	code := this.r.FormValue("code")
	state := this.r.FormValue("state")
	log.Printf("code:%s state:%s", code, state)

	// CSRF attack check
	if sess.Get(FacebookOauthState) != state {
		log.Printf("invaild state sess:%s resp:%s", sess.Get(FacebookOauthState), state)
		panic(errors.New("invaild state"))
	}

	ctx := appengine.NewContext(this.r)
	tok, err := this.config.Exchange(ctx, code)
	if err != nil {
		panic(err)
	}
	if tok.Valid() == false {
		log.Printf("invaild token:%#v", tok)
		panic(errors.New("invaild token"))
	}
	this.tok = tok
	sess.Set(FacebookAccessTokenKey, tok.AccessToken)

	fb, err := this.getMe()
	if err != nil {
		log.Printf("Get me error:%#v", err)
		panic(err)
	}
	if id, ok := fb["id"].(string); ok {
		sess.Set("id", id)
	}
	if email, ok := fb["email"].(string); ok {
		sess.Set("email", email)
	}
	if name, ok := fb["name"].(string); ok {
		sess.Set("name", name)
	}
	if picture, ok := fb["picture"].(map[string]interface{}); ok {
		if data, ok := picture["data"].(map[string]interface{}); ok {
			if url, ok := data["url"].(string); ok {
				sess.Set("picture", url)
			}
		}
	}
	sess.Save()
	http.Redirect(this.w, this.r, redirect, http.StatusFound)
}

func (this *Facebook) getMe() (map[string]interface{}, error) {
	if this.config == nil {
		return nil, errors.New("client not set")
	}
	if this.tok == nil {
		return nil, errors.New("token not set")
	}
	ctx := appengine.NewContext(this.r)
	this.client = this.config.Client(ctx, this.tok)

	v := url.Values{}
	v.Set("fields", "email,name,picture")

	resp, err := this.client.Get(FacebookApiEndpoint+"me")
	if err != nil {
		log.Printf("Get me error:%#v", err)
		panic(err)
	}
	defer resp.Body.Close()

	log.Printf("Response:%s", resp.Body)

	if resp.StatusCode >= 500 {
		return nil, errors.New("Facebook is unavailable")
	}

	if resp.StatusCode >= 400 {
		return nil, errors.New("Facebook request is invalid")
	}

	var result map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	log.Printf("Result %#v", result)
	return result, nil
}
