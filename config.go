//shareit package aims at browsing files and sharing them with others
package shareit

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit/auth"
	"github.com/scritch007/shareit/database"
	"github.com/scritch007/shareit/types"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type configSubStruct struct {
	Type   string           `json:"type"`
	Config *json.RawMessage `json:"config"`
}
type readConfiguration struct {
	RootPrefix string           `json:"root_prefix"`
	PrivateKey string           `json:"private_key"`
	HtmlPrefix string           `json:"html_prefix"`
	StaticPath string           `json:"static_path"`
	WebPort    string           `json:"web_port"`
	DbConfig   configSubStruct  `json:"database"`
	AuthConfig *json.RawMessage `json:"auth"`
}

func NewConfiguration(r *mux.Router) (resultConfig *types.Configuration) {
	var help = false
	var configFile = ""
	flag.StringVar(&configFile, "config", "", "Configuration file to use")
	flag.StringVar(&configFile, "c", "", "Configuration file to use")
	flag.BoolVar(&help, "help", false, "Display Help")
	flag.BoolVar(&help, "h", false, "Display Help")

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
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	c := new(readConfiguration)
	err = json.Unmarshal(file, c)
	if nil != err {
		fmt.Printf("Couldn't read configuration content: error was %v", err)
		os.Exit(1)
	}
	//Check the configuration
	if 0 == len(c.WebPort) {
		fmt.Println("Error: web_port should be set to a correct value")
		os.Exit(2)
	}
	staticPath, err := filepath.Abs(c.StaticPath)
	if err != nil {
		fmt.Println("Couldn't get Absolute path for %s", c.StaticPath)
		os.Exit(2)
	}

	if _, err := os.Stat(staticPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Error: the path %s, doesn't exist", staticPath)
		} else {
			fmt.Println("Error: Something went wrong when accessing to %s, %v", staticPath, err)
		}
		os.Exit(2)
	}
	rootPrefix, err := filepath.Abs(c.RootPrefix)
	if err != nil {
		fmt.Println("Couldn't get Absolute path for %s", c.StaticPath)
		os.Exit(2)
	}
	if _, err := os.Stat(rootPrefix); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Error: the path %s, doesn't exist", rootPrefix)
		} else {
			fmt.Println("Error: Something went wrong when accessing to %s, %v", rootPrefix, err)
		}
		os.Exit(2)
	}
	resultConfig = new(types.Configuration)
	resultConfig.RootPrefix = rootPrefix
	resultConfig.PrivateKey = c.PrivateKey
	resultConfig.StaticPath = staticPath
	resultConfig.WebPort = c.WebPort

	temp := path.Join(c.HtmlPrefix, "/")
	if "/" != string(temp[len(temp)-1]) {
		temp += "/"
	}

	resultConfig.HtmlPrefix = temp
	//Now Start the Auth and DB configurations...

	resultConfig.Db, err = database.NewDatabase(c.DbConfig.Type, c.DbConfig.Config)
	if nil != err {
		fmt.Println("Error: Error reading database configuration: ", err)
		os.Exit(2)
	}

	resultConfig.Auth, err = auth.NewAuthentication(c.AuthConfig, r, resultConfig)
	if nil != err {
		fmt.Println("Error: Error reading authentication configuration", err)
	}
	return resultConfig
}
