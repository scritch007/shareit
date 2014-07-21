package main

import (
	"encoding/json"
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
func (m *Main) serveBowerFiles(w http.ResponseWriter, r *http.Request) {
	m.serveFile(w, r, "bower_components/")
}

func (m *Main) serveFile(w http.ResponseWriter, r *http.Request, folder string) {
	vars := mux.Vars(r)
	file := vars["file"]
	types.LOG_DEBUG.Println("Serving file %s", file)
	http.ServeFile(w, r, path.Join(m.path, folder, file))
}

func (m *Main) homeHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if "polymer" == mode{
		http.ServeFile(w, r, path.Join(m.path, "index-polymer.html"))
	}else{
		http.ServeFile(w, r, path.Join(m.path, "index.html"))
	}
}

func (m *Main) authsHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(m.config.Auth.GetAvailableAuthentications())
	io.WriteString(w, string(b))
}

func (m *Main) serveHTMLFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]
	http.ServeFile(w, r, path.Join(m.path, "html", file+".html"))
}

type ShareMinatorConfig struct {
	AllowChangingAccesses bool `json:"change_access"`
}

func (m *Main) configHandler(w http.ResponseWriter, r *http.Request) {
	conf := ShareMinatorConfig{false}
	b, _ := json.Marshal(conf)
	io.WriteString(w, string(b))
}

func main() {
	types.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	r := mux.NewRouter()
	config := shareit.NewConfiguration(r)

	m := NewMain(config)

	c := shareit.NewCommandHandler(config)

	r.HandleFunc(config.HtmlPrefix, m.homeHandler)
	r.HandleFunc(path.Join(config.HtmlPrefix, "config"), m.configHandler)
	r.HandleFunc(path.Join(config.HtmlPrefix, "commands"), c.Commands).Methods("GET", "POST")
	r.HandleFunc(path.Join(config.HtmlPrefix, "commands/{command_id}"), c.Command).Methods("GET", "PUT", "DELETE", "POST")
	r.HandleFunc(path.Join(config.HtmlPrefix, "downloads/{file:.*}"), c.Download).Methods("GET")
	r.HandleFunc(path.Join(config.HtmlPrefix, "auths"), m.authsHandler).Methods("GET")
	r.HandleFunc(path.Join(config.HtmlPrefix, "auths/logout"), config.Auth.LogOut).Methods("GET")
	r.HandleFunc(path.Join(config.HtmlPrefix, "auths/list_users"), config.Auth.ListUsers).Methods("GET")
	r.HandleFunc(path.Join(config.HtmlPrefix, "js/{file:.*}"), m.serveJSFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "css/{file:.*}"), m.serveCSSFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "img/{file:.*}"), m.serveIMGFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "fonts/{file:.*}"), m.serveFontsFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "bower_components/{file:.*}"), m.serveBowerFiles)
	r.HandleFunc(path.Join(config.HtmlPrefix, "{file}.html"), m.serveHTMLFile)

	http.Handle(config.HtmlPrefix, r)

	types.LOG_INFO.Println("Starting server on port " + m.port)
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
