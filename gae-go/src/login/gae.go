package login

import (
	"log"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/user"
)

type GAE struct {
	w	http.ResponseWriter
	r *http.Request
}

func NewGAE(w http.ResponseWriter, r *http.Request) *GAE {
	return &GAE{w, r}
}

func (this *GAE) Login(redirect string) {
	ctx := appengine.NewContext(this.r)
	url, err := user.LoginURL(ctx, redirect)
	if err != nil {
		http.Error(this.w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(this.w, this.r, url, http.StatusFound)
}

func (this *GAE) Callback(redirect string) {
	// TODO get callback param
	ctx := appengine.NewContext(this.r)
	u, err := user.CurrentOAuth(ctx, "")
	if err != nil {
		http.Error(this.w, "OAuth Authorization header required", http.StatusUnauthorized)
		return
	}
	/*
	if !u.Admin {
		// Admin only access
		http.Error(this.w, "Admin login only", http.StatusUnauthorized)
	}
	*/
	// TODO set user to sesssion
	sess := NewSession(this.w, this.r, &sessionConfig)
	log.Printf("sess %#v", sess.session)
	sess.Set("id", u.ID)
	sess.Set("email", u.Email)
	sess.Set("name", "")
	sess.Set("photo", "")
	sess.Set("provider", u.FederatedProvider)
	sess.Save()

	http.Redirect(this.w, this.r, redirect, http.StatusFound)
}