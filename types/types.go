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
	ListAccounts(searchDict map[string]string)(accounts []*Account, err error)

	StoreSession(session *Session)(err error)
	GetSession(ref string)(session *Session, err error)
	RemoveSession(ref string)(err error)

	//ShareLink
	SaveShareLink(shareLink * ShareLinkCommand)(err error)
	GetShareLink(key string) (shareLink * ShareLinkCommand, err error)
	RemoveShareLink(key string)(err error)
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
type BrowserCommandList struct {
	Path    string        `json:"path"`
	Results []StorageItem `json:"results"`
}

//Create Folder command structure. This is used for request and response
type BrowserCommandCreateFolder struct {
	Path   string      `json:"path"`
	Result StorageItem `json:"result"`
}

type BrowserCommandDeleteItem struct {
	Path string `json:"path"`
}

type BrowserCommandDownloadLink struct {
	Path   string       `json:"path"`
	Result DownloadLink `json:"download_link"`
}


type EnumShareLinkType string

const (
	EnumShareByKey     EnumShareLinkType = "key" //EveryBody with the key can access to this sharelink
	EnumRestricted     EnumShareLinkType = "restricted" // Shared to a limited number of users can access to this link
	EnumAuthenticated  EnumShareLinkType = "authenticated" //Only the authenticated users with the key can access to this sharelink
)
type ShareLinkCommand struct{
	Path 	*string `json:"path,omitempty"` //Can be empty only if ShareLinkKey is provided
	Key *string `json:"key:omitempty` //Can be empty only for a creation
	User    string `json:"user"` //This will only be set by server. This is the user that issued the share link
	UserList *[]string `json:"user_list,omitempty"` //This is only available for EnumRestricted mode
	Type    EnumShareLinkType `json:"type"`
}

type BrowserCommand struct{
	List               	 *BrowserCommandList         `json:"list,omitempty"`
	Delete               *BrowserCommandDeleteItem   `json:"delete,omitempty"`
	CreateFolder         *BrowserCommandCreateFolder `json:"create_folder,omitempty"`
	GenerateDownloadLink *BrowserCommandDownloadLink `json:"download_link,omitempty"`
}

type Command struct {
	Name                 EnumAction           `json:"name"`       // Name of action Requested
	CommandId            string               `json:"command_id"` // Command Id returned by client when timeout is reached
	State                CommandStatus        `json:"state"`
	Timeout              int                  `json:"timeout"` // Result should be returned before timeout, or client will have to poll using CommandId
	User 				 *string              `json:"user,omitempty"`//This is only used internally to know who is actually making the request
	AuthKey              *string              `json:"auth_key,omitempty"` //Used when calling commands on behalf of a sharedlink
	ShareLink            *ShareLinkCommand    `json:"share_link,omitempty"`
	Browser               *BrowserCommand       `json:"browser,omitempty"`
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
	EnumBrowserBrowse       EnumAction = "browser.list"
	EnumBrowserCreateFolder EnumAction = "browser.create_folder"
	EnumBrowserDeleteItem   EnumAction = "browser.delete_item"
	EnumBrowserDownloadLink EnumAction = "browser.download_link"
	EnumDebugLongRequest    EnumAction = "debug.long_request"
	EnumShareLinkCreate     EnumAction = "share_link.create"
	EnumShareLinkUpdate     EnumAction = "share_link.update"
	EnumShareLinkDelete     EnumAction = "share_link.delete"
)
