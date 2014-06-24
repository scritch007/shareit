//shareit package aims at browsing files and sharing them with others
package types

import (
	"encoding/json"
	//  "fmt"
)

type DatabaseInterface interface {
	Name() string
	ListCommands(user *string, offset int, limit int, search_parameters *CommandsSearchParameters) ([]*Command, int, error) //set limit to -1 for retrieving all the elements, search_parameters will be used to filter response
	SaveCommand(command *Command)(err error)
	GetCommand(ref string) (command *Command, err error)
	DeleteCommand(ref *string) (err error)
	AddDownloadLink(link *DownloadLink) (err error)
	GetDownloadLink(path string) (link *DownloadLink, err error)

	//We need to implement some authentication structures to
	AddAccount(account *Account) (err error)
	GetAccount(authType string, ref string) (account *Account, err error)
	GetUserAccount(id string)(account *Account, err error)
	ListAccounts(map[string]string searchDict)(accounts []*Account, err error)

	StoreSession(session *Session)(err error)
	GetSession(ref string)(session *Session, err error)
	RemoveSession(ref string)(err error)
}

type Account struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	AuthType string `json:"auth_type"`
	Id       string `json:"id"` //This id should be unique depending on the DbType
	Blob     string `json:"specific"`
	//Add some other infos here for all the specific stuffs
}

type Session struct {
	AuthenticationHeader string
	UserId string
}

//Configuration Structure
type Configuration struct {
	RootPrefix string
	PrivateKey string
	StaticPath string
	WebPort    string
	Db         DatabaseInterface
	Auth 	   *Authentication
}

type DownloadLink struct {
	Link string `json:"link"`
	Path string `json:"path"`
}

func (d *DownloadLink) String() string {
	b, err := json.Marshal(d)
	if nil != err {
		return "Couldn't serialize"
	}
	return string(b)
}

//Current Status of the command
type EnumStatus int

const (
	COMMAND_STATUS_DONE        EnumStatus = 0
	COMMAND_STATUS_QUEUED      EnumStatus = 1
	COMMAND_STATUS_IN_PROGRESS EnumStatus = 2
	COMMAND_STATUS_ERROR       EnumStatus = 3
	COMMAND_STATUS_CANCELLED   EnumStatus = 4
)

//Defines current command status + error code and current Progress
type CommandStatus struct {
	Status    EnumStatus `json:"status"`
	Progress  int        `json:"progress,omitempty"`
	ErrorCode int        `json:"error_code,omitempty"`
}

//Element describing the an item
type StorageItem struct {
	Name             string `json:"name"`
	IsDir            bool   `json:"isDir"`
	ModificationDate int64  `json:"mDate"`
	Size             int64  `json:"size"`
	Kind             string `json:"kind"`
}
type CommandsSearchParameters struct {
	Status *EnumStatus `json:"status,omitempty"`
}

//Browse command structure. This is used for request and response
type CommandBrowse struct {
	Path    string        `json:"path"`
	Results []StorageItem `json:"results"`
}

//Create Folder command structure. This is used for request and response
type CommandCreateFolder struct {
	Path   string      `json:"path"`
	Result StorageItem `json:"result"`
}

type CommandDeleteItem struct {
	Path string `json:"path"`
}

type CommandDownloadLink struct {
	Path   string       `json:"path"`
	Result DownloadLink `json:"download_link"`
}

type Command struct {
	Name                 EnumAction           `json:"name"`       // Name of action Requested
	CommandId            string               `json:"command_id"` // Command Id returned by client when timeout is reached
	State                CommandStatus        `json:"state"`
	Timeout              int                  `json:"timeout"` // Result should be returned before timeout, or client will have to poll using CommandId
	Browse               *CommandBrowse       `json:"browse_command,omitempty"`
	Delete               *CommandDeleteItem   `json:"delete_command,omitempty"`
	CreateFolder         *CommandCreateFolder `json:"create_folder_command,omitempty"`
	GenerateDownloadLink *CommandDownloadLink `json:"download_link_command,omitempty"`
	User 				 *string              //This is only used internally to know who is actually making the request
}

func (c *Command) String() string {
	b, err := json.Marshal(c)
	if nil != err {
		return "Couldn't serialize"
	}
	return string(b)
}

type EnumAction string

const (
	EnumBrowserBrowse       EnumAction = "browser.browse"
	EnumBrowserCreateFolder EnumAction = "browser.create_folder"
	EnumBrowserDeleteItem   EnumAction = "browser.delete_item"
	EnumBrowserDownloadLink EnumAction = "browser.download_link"
	EnumDebugLongRequest    EnumAction = "debug.long_request"
)
