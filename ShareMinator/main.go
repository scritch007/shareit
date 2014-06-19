package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit"
	"github.com/scritch007/shareit/types"
	"io"
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
	types.LOG_DEBUG.Println("Serving file %s", file)
	http.ServeFile(w, r, path.Join(m.path, folder, file))
}

func (m *Main) homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got home request")
	http.ServeFile(w, r, path.Join(m.path, "index.html"))
}

func (m *Main) authsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Requested the available auths")
	b, _ := json.Marshal(m.config.GetAvailableAuthentications())
	io.WriteString(w, string(b))
}

func main() {
	types.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	r := mux.NewRouter()
	config := shareit.NewConfiguration(r)

	m := NewMain(config)

	c := shareit.NewCommandHandler(config)

	r.HandleFunc("/", m.homeHandler)
	r.HandleFunc("/commands", c.Commands).Methods("GET", "POST")
	r.HandleFunc("/commands/{command_id}", c.Command).Methods("GET", "PUT", "DELETE")
	r.HandleFunc("/downloads/{file:.*}", c.Download).Methods("GET")
	r.HandleFunc("/auths", m.authsHandler).Methods("GET")
	r.HandleFunc("/js/{file:.*}", m.serveJSFile)
	r.HandleFunc("/css/{file:.*}", m.serveCSSFile)
	r.HandleFunc("/img/{file:.*}", m.serveIMGFile)
	r.HandleFunc("/fonts/{file:.*}", m.serveFontsFile)

	http.Handle("/", r)

	fmt.Println("Starting server on port " + m.port)
	http.ListenAndServe(":"+m.port, nil)
}

type Main struct {
	path   string
	port   string
	config *types.Configuration
}

func NewMain(configuration *types.Configuration) (m *Main) {
	m = new(Main)
	m.path = configuration.StaticPath
	m.port = configuration.WebPort
	m.config = configuration
	return m
}
