package template

import (
"html/template"
"log"
"net/http"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
		Body  string
	}{
		Title: "default title",
		Body:  "default page body",
	}

	execute(w, template.Must(template.ParseFiles(
		"view/layout.html",
		"view/index.html")), "layout.html", data)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
		Body  string
	}{
		Title: "page title",
		Body:  "page body",
	}

	execute(w, template.Must(template.ParseFiles(
		"view/layout.html",
		"view/page.html")), "layout.html", data)
}

func execute(w http.ResponseWriter, t *template.Template, name string, data interface{}) {
	// テンプレートを描画
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Fatal(err)
	}
}

func init() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/page/", pageHandler)
}