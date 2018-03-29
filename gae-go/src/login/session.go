package login

import (
	"net/http"
	"log"

	"github.com/gorilla/sessions"
)

type Session struct {
	w	http.ResponseWriter
	r *http.Request
	store *sessions.CookieStore
	session *sessions.Session
}

type SessionConfig struct {
	Name string
	CookieSecret string
	Domain string
	Path string
	MaxAge int
	Secure bool
	HttpOnly bool
}

func NewSession(w http.ResponseWriter, r *http.Request, c *SessionConfig) *Session {
	store := sessions.NewCookieStore([]byte(c.CookieSecret))
	if c != nil {
		store.Options = &sessions.Options{
			Domain:     c.Domain,
			Path:       c.Path,
			MaxAge:     c.MaxAge,
			Secure:     c.Secure,
			HttpOnly:   c.HttpOnly,
		}
	}
	s, err := store.Get(r, c.Name)
	if err != nil {
		log.Printf("sess no data %s %#v", c.Name, err)
		s, err = store.New(r, c.Name)
	}
	sess := &Session{w, r, store, s}
	log.Printf("sess create Name:%s SessionKey:%s ID:%s", c.Name, c.CookieSecret, s.ID)
	return sess
}

// TODO
func destroySession(w http.ResponseWriter, r *http.Request) {

}

func (this *Session) Set(k string, v string) {
	this.session.Values[k] = v
}

func (this *Session) Get(k string) string {
	if v := this.session.Values[k]; v != nil {
		return v.(string)
	}
	return ""
}

func (this *Session) Save() {
	err := this.store.Save(this.r, this.w, this.session)
	log.Printf("sess save %#v", this.session)
	if err != nil {
		log.Printf("sess err %#v", err)
	}
}
