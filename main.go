package main

import (
	"github.com/gorilla/mux"
    "fmt"
    "net/http"
)

func (m *Main)serveJSFile(w http.ResponseWriter, r *http.Request){
	fmt.Println("Got request for js file");
	vars := mux.Vars(r)
    file := vars["file"]
	http.ServeFile(w, r, m.path + "js/" + string(file))
}

func (m *Main)serveCSSFile(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
    file := vars["file"]
	http.ServeFile(w, r, m.path + "css/" + string(file))
}

func (m *Main)homeHandler(w http.ResponseWriter, r *http.Request){
	fmt.Println("Got home request");
	http.ServeFile(w, r, m.path + "index.html")
}

func main() {
	m := new(Main)
	m.init()

	config := new(Configuration)
	config.init()

	c := new(CommandHandler)
	c.Init(config)

    r := mux.NewRouter()
    r.HandleFunc("/", m.homeHandler)
    r.HandleFunc("/commands", c.Commands).Methods("GET", "POST")
    r.HandleFunc("/commands/{command_id}", c.Command).Methods("GET", "PUT", "DELETE")
    r.HandleFunc("/js/{file}", m.serveJSFile)
    r.HandleFunc("/css/{file}", m.serveCSSFile)
    http.Handle("/", r)

    fmt.Println("Starting server on port " + m.port)
    http.ListenAndServe(":"+m.port, nil)
}


type Main struct{
	path string
	port string
}

func (m *Main)init(){
	m.path = "./html/"
	m.port = "8080"
}

type Configuration struct{
	RootPrefix string
	DbPath string
	DbUser string
	DbPassword string
}

func (c *Configuration)init(){
	c.RootPrefix = "/home/benjamin/dev/go"
	return
}