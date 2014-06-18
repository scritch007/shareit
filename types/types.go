//shareit package aims at browsing files and sharing them with others
package types

import (
	"github.com/gorilla/mux"
)

type DatabaseInterface interface {
	Name() string
	AddCommand(command *Command) (ref string, err error)
	ListCommands(offset int, limit int, search_parameters *CommandsSearchParameters) ([]Command, int, error) //set limit to -1 for retrieving all the elements, search_parameters will be used to filter response
	GetCommand(ref string)(command *Command, err error)
	DeleteCommand(ref *string, err error)
	AddDownloadLink(key string, link *DownloadLink)(err error)
	GetDownloadLink(path string)(link *DownloadLink, err error)
}

type Authentication interface {
	Name() string
	AddRoutes(r *mux.Router) error
}

//Configuration Structure
type Configuration struct {
	RootPrefix string
	PrivateKey string
	StaticPath string
	WebPort    string
	Db         DatabaseInterface
	Auths      []Authentication
}

func (c *Configuration) GetAvailableAuthentications() []string {
	res := make([]string, len(c.Auths))
	for i, elem := range c.Auths {
		res[i] = elem.Name()
	}
	return res
}

type DownloadLink struct{
	Link string `json:"link"`
	Path string `json:"path"`
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
	ChangeDate       int    `json:"cDate"`
	Size             int64  `json:"size"`
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
	Path   string `json:"path"`
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
	GenerateDownloadLink *CommandDownloadLink `json:"download_link_command"`
}

type EnumAction string

const (
	EnumBrowserBrowse       EnumAction = "browser.browse"
	EnumBrowserCreateFolder EnumAction = "browser.create_folder"
	EnumBrowserDeleteItem   EnumAction = "browser.delete_item"
	EnumBrowserDownloadLink EnumAction = "browser.download_link"
	EnumDebugLongRequest    EnumAction = "debug.long_request"
)