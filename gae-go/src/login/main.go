package login

import (
	"html/template"
	"os"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"strconv"
)

const (
)

type TemplateData struct {
	Title string
	Body  string
	Options map[string]interface{}
}

var (
	sessionConfig = SessionConfig{}
	templateData = TemplateData{"site name", "site body", map[string]interface{}{
		"enableGoogle": false,
		"enableFacebook": false,
		"enableTwitter": false,
	}}
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("templateData %#v", templateData)
	execute(w, template.Must(template.ParseFiles(
		"view/layout.html",
		"view/login.html")), "layout.html", templateData)
}

func memberHandler(w http.ResponseWriter, r *http.Request) {
	templateData = TemplateData{"member", "member body", map[string]interface{}{}}

	execute(w, template.Must(template.ParseFiles(
		"view/layout.html",
		"view/member.html")), "layout.html", templateData)
}

func memberDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	templateData = TemplateData{"member detail", "member body", map[string]interface{}{}}

	if id == "me" {
		// self
		sess := NewSession(w, r, &sessionConfig)
		if sess == nil {
			panic("error session get")
		}
		log.Printf("sess show %#v", sess.session)
		templateData.Title = "member detail:"+sess.Get("id")
		templateData.Body = "member detail body "+sess.Get("id")+sess.Get("email")+sess.Get("provider")
	} else {
		templateData.Title = "member detail:"+id
		templateData.Body = "member detail body "+id
	}
	execute(w, template.Must(template.ParseFiles(
		"view/layout.html",
		"view/member_detail.html")), "layout.html", templateData)
}

func execute(w http.ResponseWriter, t *template.Template, n string, d interface{}) {
	// テンプレートを描画
	if err := t.ExecuteTemplate(w, n, d); err != nil {
		log.Fatal(err)
	}
}

func init() {
	sessionConfig = SessionConfig{
		Name: os.Getenv("SESSION_NAME"),
		CookieSecret: os.Getenv("SESSION_COOKIE_SECRET"),
		Domain: os.Getenv("SESSION_DOMAIN"),
		Path: os.Getenv("SESSION_PATH"),
	}
	sessionConfig.MaxAge, _ = strconv.Atoi(os.Getenv("SESSION_MAXAGE"))
	sessionConfig.Secure, _ = strconv.ParseBool(os.Getenv("SESSION_SECURE"))
	sessionConfig.HttpOnly, _ = strconv.ParseBool(os.Getenv("SESSION_HTTPONLY"))

	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler).Methods(http.MethodGet)

	memberRouter := router.PathPrefix("/member").Subrouter()
	memberRouter.HandleFunc("", memberHandler).Methods(http.MethodGet)
	memberRouter.HandleFunc("/", memberHandler).Methods(http.MethodGet)
	memberRouter.HandleFunc("/{id}", memberDetailHandler).Methods(http.MethodGet)
	memberRouter.HandleFunc("/{id}/", memberDetailHandler).Methods(http.MethodGet)

	loginRouter := router.PathPrefix("/login").Subrouter()

	// GAE UserAPI
	loginRouter.HandleFunc("/gae", func(w http.ResponseWriter, r *http.Request) {
		g := NewGAE(w, r)
		g.Login("/login/gae/callback")
	}).Methods(http.MethodGet)

	loginRouter.HandleFunc("/gae/callback", func(w http.ResponseWriter, r *http.Request) {
		g := NewGAE(w, r)
		g.Callback("/member/me/")
	}).Methods(http.MethodGet)

	// Google Oauth
	if os.Getenv("GOOGLE_CLIENT_ID") != "" &&
 		 os.Getenv("GOOGLE_CLIENT_SECRET") != "" {
 		templateData.Options["enableGoogle"] = true
		loginRouter.HandleFunc("/google", func(w http.ResponseWriter, r *http.Request) {
			g := NewGoogle(w, r)
			g.Login()
		}).Methods(http.MethodGet)

		loginRouter.HandleFunc("/google/callback", func(w http.ResponseWriter, r *http.Request) {
			g := NewGoogle(w, r)
			g.Callback("/member/me/")
		}).Methods(http.MethodGet)
	}

	// Twitter Oauth
	if os.Getenv("TWITTER_CONSUMER_KEY") != "" &&
		os.Getenv("TWITTER_CONSUMER_SECRET") != "" {
		templateData.Options["enableTwitter"] = true
		loginRouter.HandleFunc("/twitter", func(w http.ResponseWriter, r *http.Request) {
			tw := NewTwitter(w, r)
			tw.Login()
		}).Methods(http.MethodGet)

		loginRouter.HandleFunc("/twitter/callback", func(w http.ResponseWriter, r *http.Request) {
			tw := NewTwitter(w, r)
			tw.Callback("/member/me/")
		}).Methods(http.MethodGet)

		loginRouter.HandleFunc("/twitter/me", func(w http.ResponseWriter, r *http.Request) {
			tw := NewTwitter(w, r)
			_, err := tw.getMe()
			if err != nil {
				http.Redirect(w, r, "/login/twitter", http.StatusFound)
			}
		}).Methods(http.MethodGet)
	}

	// Facebook Oauth
	// Twitter Oauth
	if os.Getenv("FACEBOOK_APP_ID") != "" &&
		os.Getenv("FACEBOOK_APP_SECRET") != "" {
		templateData.Options["enableFacebook"] = true
		loginRouter.HandleFunc("/facebook", func(w http.ResponseWriter, r *http.Request) {
			fb := NewFacebook(w, r)
			fb.Login()
		}).Methods(http.MethodGet)

		loginRouter.HandleFunc("/facebook/callback", func(w http.ResponseWriter, r *http.Request) {
			fb := NewFacebook(w, r)
			fb.Callback("/member/me/")
		}).Methods(http.MethodGet)

		loginRouter.HandleFunc("/facebook/me", func(w http.ResponseWriter, r *http.Request) {
			fb := NewFacebook(w, r)
			_, err := fb.getMe()
			if err != nil {
				http.Redirect(w, r, "/login/facebook", http.StatusFound)
			}
		}).Methods(http.MethodGet)
	}

	//router.NotFoundHandler = http.HandlerFunc(indexHandler)

	http.Handle("/", router)

}
