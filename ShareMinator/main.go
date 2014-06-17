package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit"
	"net/http"
	"os"
	"path"
)

func (m *Main) serveJSFile(w http.ResponseWriter, r *http.Request) {
	m.serveFile(w, r, "js/")
}

func (m *Main) serveCSSFile(w http.ResponseWriter, r *http.Request) {
	m.serveFile(w, r, "css/")
}

func (m *Main) serveIMGFile(w http.ResponseWriter, r *http.Request) {
	m.serveFile(w, r, "img/")
}
func (m *Main) serveFontsFile(w http.ResponseWriter, r *http.Request) {
	m.serveFile(w, r, "fonts/")
}

func (m *Main) serveFile(w http.ResponseWriter, r *http.Request, folder string) {
	vars := mux.Vars(r)
	file := vars["file"]
	shareit.LOG_DEBUG.Println("Serving file %s", file)
	http.ServeFile(w, r, path.Join(m.path, folder, file))
}

func (m *Main) homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got home request")
	http.ServeFile(w, r, path.Join(m.path, "index.html"))
}

func main() {
	shareit.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	config := shareit.NewConfiguration()

	m := NewMain(config)

	c := shareit.NewCommandHandler(config)

	a := shareit.NewAuthentication(config)

	r := mux.NewRouter()
	r.HandleFunc("/", m.homeHandler)
	r.HandleFunc("/commands", c.Commands).Methods("GET", "POST")
	r.HandleFunc("/commands/{command_id}", c.Command).Methods("GET", "PUT", "DELETE")
	r.HandleFunc("/downloads/{file:.*}", c.Download).Methods("GET")
	r.HandleFunc("/js/{file:.*}", m.serveJSFile)
	r.HandleFunc("/css/{file:.*}", m.serveCSSFile)
	r.HandleFunc("/img/{file:.*}", m.serveIMGFile)
	r.HandleFunc("/fonts/{file:.*}", m.serveFontsFile)

	r.HandleFunc("/auth/{method:.*}", a.Handle)

	http.Handle("/", r)

	fmt.Println("Starting server on port " + m.port)
	http.ListenAndServe(":"+m.port, nil)
}

type Main struct {
	path   string
	port   string
	config *shareit.Configuration
}

func NewMain(configuration *shareit.Configuration) (m *Main) {
	m = new(Main)
	m.path = configuration.StaticPath
	m.port = configuration.WebPort
	m.config = configuration
	return m
}
