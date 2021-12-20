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

	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit"
	"github.com/scritch007/shareit/types"

	echo "github.com/labstack/echo/v4"
)

type Theme struct {
	Name string
}

func (m *Main) homeHandler(ctx echo.Context) error {
	r := ctx.Request()

	mode := r.URL.Query().Get("mode")
	if mode == "polymer" {
		return ctx.File(path.Join(m.path, "index-polymer.html"))
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
			return fmt.Errorf("failed to get template %w", err)
		}
		err = t.Execute(ctx.Response(), Theme{Name: file})
		if nil != err {
			return fmt.Errorf("failed to write template %w", err)
		}
	}
	return nil
}

func (m *Main) authsHandler(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(m.config.Auth.GetAvailableAuthentications())
	io.WriteString(w, string(b))
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

	if len(configFile) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	if help {
		flag.Usage()
		os.Exit(0)
	}

	e := echo.New()
	e.Debug = true

	config := shareit.NewConfiguration(configFile, e)

	m := NewMain(config)

	c := shareit.NewCommandHandler(config)

	e.Any(path.Join(config.HtmlPrefix, "config"), func(ctx echo.Context) error {
		m.configHandler(ctx.Response(), ctx.Request())
		return nil
	})

	e.Any(path.Join(config.HtmlPrefix, "commands"), func(ctx echo.Context) error {
		c.Commands(ctx.Response(), ctx.Request())
		return nil
	}) //.Methods("GET", "POST")
	e.Any(path.Join(config.HtmlPrefix, "commands/:command_id"), func(ctx echo.Context) error {
		c.Command(ctx.Response(), ctx.Request())
		return nil
	}) //.Methods("GET", "PUT", "DELETE", "POST")
	e.Any(path.Join(config.HtmlPrefix, "downloads/:file"), c.Download) //.Methods("GET")
	e.Any(path.Join(config.HtmlPrefix, "auths"), func(ctx echo.Context) error {
		m.authsHandler(ctx.Response(), ctx.Request())
		return nil
	}) //.Methods("GET")
	e.Any(path.Join(config.HtmlPrefix, "auths/logout"), func(ctx echo.Context) error {
		config.Auth.LogOut(ctx.Response(), ctx.Request())
		return nil
	}) //.Methods("GET")
	e.Any(path.Join(config.HtmlPrefix, "auths/list_users"), func(ctx echo.Context) error {
		config.Auth.ListUsers(ctx.Response(), ctx.Request())
		return nil
	}) //.Methods("GET")
	e.Any(path.Join(config.HtmlPrefix, "auths/get_info"), func(ctx echo.Context) error {
		config.Auth.GetInfo(ctx.Response(), ctx.Request())
		return nil
	}) //.Methods("GET")

	e.GET(config.HtmlPrefix, m.homeHandler)
	e.Static(config.HtmlPrefix, config.StaticPath)

	corsObj := handlers.AllowedOrigins([]string{"*"})
	methodObjs := handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodOptions})
	headersObjs := handlers.AllowedHeaders([]string{"Content-Type", "Accept", "Accept-Language", "Content-Language", "Origin"})

	s := &http.Server{Addr: ":" + m.port, Handler: handlers.CORS(corsObj, methodObjs, headersObjs)(e)}

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
