package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"errors"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
	Title string
	Body []byte
}

func getTitle(response http.ResponseWriter, request *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(request.URL.Path)
	if m == nil {
		http.NotFound(response, request)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{ Title: title, Body: body }, nil
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		m := validPath.FindStringSubmatch(request.URL.Path)
		if m == nil {
			http.NotFound(response, request)
			return
		}
		fn(response, request, m[2])
	}
}

func renderTemplate(response http.ResponseWriter, tmpl string, page *Page) {
	err := templates.ExecuteTemplate(response, tmpl+".html", page)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(response http.ResponseWriter, request *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(response, request, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(response, "view", page)
}

func editHandler(response http.ResponseWriter, request *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderTemplate(response, "edit", page)
}

func saveHandler(response http.ResponseWriter, request *http.Request, title string) {
	body := request.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(response, request, "/view/"+title, http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}
