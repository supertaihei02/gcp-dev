package login

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

type Session struct {
	w	http.ResponseWriter
	r *http.Request
	store *sessions.CookieStore
	session *sessions.Session
}

func NewSession(w http.ResponseWriter, r *http.Request) *Session {
	store := sessions.NewCookieStore([]byte("something-very-secret"))
	s, err := store.Get(r, SessionName)
	if err != nil {
		log.Printf("sess no data %s %#v", SessionName, err)
		s, err = store.New(r, SessionName)
	}
	context.Set(r, ContextSessionKey, s)
	log.Printf("sess start Name:%s SessionKey:%s ID:%s", SessionName, ContextSessionKey, s.ID)
	return &Session{w, r, store, s}
}

func (this *Session) Flush() {
	if s := context.Get(this.r, ContextSessionKey).(*sessions.Session); s != nil {
		this.session = s
	}
}

func (this *Session) Set(k string, v string) {
	this.session.Values[k] = v
}

func (this *Session) Get(k string) string {
	if v := this.session.Values[k]; v != nil {
		return v.(string)
	}
	return "no data"
}

func (this *Session) Save() {
	err := this.store.Save(this.r, this.w, this.session)
	log.Printf("sess save %#v", this.session)
	if err != nil {
		log.Printf("sess err %#v", err)
	}
}
