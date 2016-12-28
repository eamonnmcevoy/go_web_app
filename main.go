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

func renderTemplate(response http.ResponseWriter, tmpl string, page *Page) {
	err := templates.ExecuteTemplate(response, tmpl+".html", page)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(response http.ResponseWriter, request *http.Request) {
	title, err := getTitle(response, request)
	if err != nil {
		return
	}
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(response, request, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(response, "view", page)
}

func editHandler(response http.ResponseWriter, request *http.Request) {
	title, err := getTitle(response, request)
	if err != nil {
		return
	}
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderTemplate(response, "edit", page)
}

func saveHandler(response http.ResponseWriter, request *http.Request) {
	title, err := getTitle(response, request)
	if err != nil {
		return
	}
	body := request.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err = page.save()
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(response, request, "/view/"+title, http.StatusFound)
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.ListenAndServe(":8080", nil)
}
