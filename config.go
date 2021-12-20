//shareit package aims at browsing files and sharing them with others
package shareit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/shareit/auth"
	"github.com/scritch007/shareit/auth/dummy"
	"github.com/scritch007/shareit/database"
	"github.com/scritch007/shareit/types"
)

type configSubStruct struct {
	Type   string           `json:"type"`
	Config *json.RawMessage `json:"config"`
}

type rootUserConfig struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type access struct {
	Name   string         `json:"name"`
	Access api.AccessType `json:"access"`
}

type userAccesses struct {
	User     *string  `json:"user"` //No user name means public access
	Accesses []access `json:"accesses"`
}

type readConfiguration struct {
	RootPrefix            string           `json:"root_prefix"`
	PrivateKey            string           `json:"private_key"`
	HtmlPrefix            string           `json:"html_prefix"`
	StaticPath            string           `json:"static_path"`
	WebPort               string           `json:"web_port"`
	DbConfig              configSubStruct  `json:"database"`
	AuthConfig            *json.RawMessage `json:"auth"`
	AllowRootWrite        bool             `json:allow_root_write` //Can we create file/folder at the root
	RootUser              *rootUserConfig  `json:"root_user"`      //Used for the admin config. If not specified, then noone will be allowed to change the configuration
	UserAccesses          *[]userAccesses  `json:"user_accesses"`  //Can be empty if allow_changing_accesses is set to true. Otherwise should be set
	AllowChangingAccesses bool             `json:"allow_changing_accesses"`
	UploadChunkSize       int64            `json:"upload_chunk_size"`
	Debug                 bool             `json:"debug"`
	Public                bool             `json:"public"`
}

func NewConfiguration(configFile string, r *echo.Echo) (resultConfig *types.Configuration) {
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
	if len(c.WebPort) == 0 {
		fmt.Println("Error: web_port should be set to a correct value")
		os.Exit(2)
	}
	staticPath, err := filepath.Abs(c.StaticPath)
	if err != nil {
		fmt.Printf("Couldn't get Absolute path for %s\n", c.StaticPath)
		os.Exit(2)
	}

	if _, err := os.Stat(staticPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: the path %s, doesn't exist", staticPath)
		} else {
			fmt.Printf("Error: Something went wrong when accessing to %s, %v", staticPath, err)
		}
		os.Exit(2)
	}
	rootPrefix, err := filepath.Abs(c.RootPrefix)
	if err != nil {
		fmt.Printf("Couldn't get Absolute path for %s", c.StaticPath)
		os.Exit(2)
	}
	if _, err := os.Stat(rootPrefix); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: the path %s, doesn't exist", rootPrefix)
		} else {
			fmt.Printf("Error: Something went wrong when accessing to %s, %v", rootPrefix, err)
		}
		os.Exit(2)
	}
	resultConfig = new(types.Configuration)
	resultConfig.RootPrefix = rootPrefix
	resultConfig.PrivateKey = c.PrivateKey
	resultConfig.StaticPath = staticPath
	resultConfig.WebPort = c.WebPort
	resultConfig.UploadChunkSize = c.UploadChunkSize
	resultConfig.Debug = c.Debug

	temp := path.Join(c.HtmlPrefix, "/")
	if string(temp[len(temp)-1]) != "/" {
		temp += "/"
	}

	resultConfig.HtmlPrefix = temp
	//Now Start the Auth and DB configurations...

	resultConfig.Db, err = database.NewDatabase(c.DbConfig.Type, c.DbConfig.Config, resultConfig.Debug)
	if nil != err {
		fmt.Println("Error: Error reading database configuration: ", err)
		os.Exit(2)
	}

	resultConfig.Auth, err = auth.NewAuthentication(c.AuthConfig, r, resultConfig)
	if nil != err {
		fmt.Println("Error: Error reading authentication configuration", err)
		os.Exit(3)
	}

	//Now create the root account if if doesn't exist
	if nil != c.RootUser {
		account, id, err := resultConfig.Db.GetAccount(dummy.Name, c.RootUser.Email)
		if nil != err {
			//This means we don't have any account
			account := new(types.Account)
			account.Auths = make(map[string]types.AccountSpecificAuth)
			account.Login = c.RootUser.Login
			account.Email = c.RootUser.Email
			account.IsAdmin = true
			authSpecific := types.AccountSpecificAuth{AuthType: dummy.Name, Blob: c.RootUser.Password}
			account.Auths[dummy.Name] = authSpecific
			//TODO This should be the sha1 from the password
			err = resultConfig.Db.AddAccount(account)
			if nil != err {
				fmt.Println("Failed to create the root account")
				os.Exit(4)
			}
		} else {
			if !account.IsAdmin {
				account.IsAdmin = true
				err = resultConfig.Db.UpdateAccount(id, account)
				if nil != err {
					fmt.Println("Failed to update the root account")
					os.Exit(4)
				}
			}
		}

	}

	if !c.AllowChangingAccesses {

		if nil == c.UserAccesses {
			fmt.Println("Error: allow_changing_accesses is false and no accesses defined")
			os.Exit(4)
		}
		for _, elem := range *c.UserAccesses {
			user_id := ""
			if elem.User != nil && *elem.User != "" {
				_, user_id, err = resultConfig.Db.GetAccount(dummy.Name, *elem.User)
				if err != nil {
					fmt.Println("Error: Account not found", *elem.User)
					continue
				}
			}
			for _, access := range elem.Accesses {
				resultConfig.Db.SetAccess(&user_id, access.Name, access.Access)
			}
		}
	}

	resultConfig.AllowRootWrite = c.AllowRootWrite
	resultConfig.Public = c.Public
	return resultConfig
}
