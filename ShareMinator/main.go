package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime/pprof"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit"
	"github.com/scritch007/shareit/types"
)

func (m *Main) serveJSFile(w http.ResponseWriter, r *http.Request) {
	m.serveFile(w, r, "js/")
}

func (m *Main) serveCSSFile(w http.ResponseWriter, r *http.Request) {
	m.serveFile(w, r, "css")
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
	tools.LOG_DEBUG.Printf("Serving file %s\n", file)
	http.ServeFile(w, r, path.Join(m.path, folder, file))
}

type Theme struct {
	Name string
}

func (m *Main) homeHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if "polymer" == mode {
		http.ServeFile(w, r, path.Join(m.path, "index-polymer.html"))
	} else {
		cookie, err := r.Cookie("theme")
		var file string
		if nil == err {
			file = "bootstrap." + cookie.Value + ".min.css"
		} else {
			file = "bootstrap.main.min.css"
		}
		t, err := template.ParseFiles(path.Join(m.path, "index.html"))
		if nil != err {
			io.WriteString(w, "Failed to get template")
		}
		err = t.Execute(w, Theme{Name: file})
		if nil != err {
			io.WriteString(w, "Failed to write template "+err.Error())
		}
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
	tools.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	r := mux.NewRouter()

	var help = false
	var configFile = ""
	flag.StringVar(&configFile, "config", "", "Configuration file to use")
	flag.StringVar(&configFile, "c", "", "Configuration file to use")
	flag.BoolVar(&help, "help", false, "Display Help")
	flag.BoolVar(&help, "h", false, "Display Help")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if 0 == len(configFile) {
		flag.Usage()
		os.Exit(0)
	}

	if help {
		flag.Usage()
		os.Exit(0)
	}

	config := shareit.NewConfiguration(configFile, r)

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
	r.HandleFunc(path.Join(config.HtmlPrefix, "auths/get_info"), config.Auth.GetInfo).Methods("GET")
	r.HandleFunc(path.Join(config.HtmlPrefix, "js/{file:.*}"), m.serveJSFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "css/{file:.*}"), m.serveCSSFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "img/{file:.*}"), m.serveIMGFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "fonts/{file:.*}"), m.serveFontsFile)
	r.HandleFunc(path.Join(config.HtmlPrefix, "bower_components/{file:.*}"), m.serveBowerFiles)
	r.HandleFunc(path.Join(config.HtmlPrefix, "{file}.html"), m.serveHTMLFile)

	corsObj := handlers.AllowedOrigins([]string{"*"})
	methodObjs := handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodOptions})
	headersObjs := handlers.AllowedHeaders([]string{"Content-Type", "Accept", "Accept-Language", "Content-Language", "Origin"})

	s := &http.Server{Addr: ":" + m.port, Handler: handlers.CORS(corsObj, methodObjs, headersObjs)(r)}

	tools.LOG_INFO.Println("Starting server on port " + m.port)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			tools.LOG_ERROR.Println("Failed to create profiling file with error " + err.Error())
			os.Exit(-1)
		}
		pprof.StartCPUProfile(f)
	}

	go func() {
		<-sig
		tools.LOG_INFO.Println("Got a shutdown command. Going down...")
		if *cpuprofile != "" {
			pprof.StopCPUProfile()
		}
		os.Exit(0)
	}()

	s.ListenAndServe()
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
