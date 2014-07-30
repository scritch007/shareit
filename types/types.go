package types

import (
	"encoding/json"
	"net/http"
	//  "fmt"
)

type DatabaseInterface interface {
	Name() string
	ListCommands(user *string, offset int, limit int, search_parameters *CommandsSearchParameters) ([]*Command, int, error) //set limit to -1 for retrieving all the elements, search_parameters will be used to filter response
	SaveCommand(command *Command) (err error)
	GetCommand(ref string) (command *Command, err error)
	DeleteCommand(ref *string) (err error)
	AddDownloadLink(link *DownloadLink) (err error)
	GetDownloadLink(path string) (link *DownloadLink, err error)

	//We need to implement some authentication structures to
	AddAccount(account *Account) (err error)
	GetAccount(authType string, ref string) (account *Account, id string, err error) // if authType = "" we just want to match for the id, do not care about the authType
	UpdateAccount(id string, account *Account) (err error)
	GetUserAccount(id string) (account *Account, err error)
	ListAccounts(searchDict map[string]string) (accounts []*Account, err error)

	GetAccess(user *string, path string) (AccessType, error)      //If user is nil then it's public access
	SetAccess(user *string, path string, access AccessType) error // This can only be called by a admin...
	ClearAccesses() error                                         //Should only be called when configuration is done by files

	StoreSession(session *Session) (err error)
	GetSession(ref string) (session *Session, err error)
	RemoveSession(ref string) (err error)

	//ShareLink
	SaveShareLink(shareLink *ShareLink) (err error)
	GetShareLink(key string) (shareLink *ShareLink, err error)
	ListShareLinks(user string) (shareLinks []*ShareLink, err error) //List the sharelink user created
	GetShareLinksFromPath(path string, user string) (shareLinks []*ShareLink, err error)
	RemoveShareLink(key string) (err error)
}

type AccountSpecificAuth struct {
	AuthType string `json:"auth_type"`
	Blob     string `json:"specific"`
}
type Account struct {
	Login   string `json:"login"`
	Email   string `json:"email"`
	Id      string `json:"id"`       //This id should be unique depending on the DbType
	IsAdmin bool   `json:"is_admin"` //Can only be changed by the admin
	//Add some other infos here for all the specific stuffs
	Auths map[string]AccountSpecificAuth
}

type Session struct {
	AuthenticationHeader string
	UserId               string
}

//Configuration Structure
type Configuration struct {
	RootPrefix     string
	PrivateKey     string
	StaticPath     string
	HtmlPrefix     string
	WebPort        string
	AllowRootWrite bool
	Db             DatabaseInterface
	Auth           *Authentication
}

type DownloadLink struct {
	Link     string  `json:"link"`
	Path     string  `json:"path"`
	RealPath *string `json:"real_path,omitempty"` //Will only be used when storing in the DB
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

type EnumCommandErrorCode int

const (
	ERROR_NO_ERROR = iota
	ERROR_MISSING_COMMAND_BODY
	ERROR_MISSING_PARAMETERS
	ERROR_INVALID_PARAMETERS
	ERROR_NOT_ALLOWED  //Should be set when accessing to things that user should be accessing to
	ERROR_INVALID_PATH //Should be used for everything that refers to a filesystem path
	ERROR_FILE_SYSTEM  //Should be used when an action on the file system fails (listing, removing, stat)
	ERROR_SAVING       //Should only be raised to say saving failed...
	ERROR_UNKNOWN      //Nothing specific there...
)

//Defines current command status + error code and current Progress
type CommandStatus struct {
	Status    EnumStatus           `json:"status"`
	Progress  int                  `json:"progress,omitempty"`
	ErrorCode EnumCommandErrorCode `json:"error_code,omitempty"`
}

type AccessType int

const (
	NONE       AccessType = 0
	READ       AccessType = 1
	READ_WRITE AccessType = 2
)

//Element describing the an item
type StorageItem struct {
	Name             string     `json:"name"`
	IsDir            bool       `json:"isDir"`
	ModificationDate int64      `json:"mDate"`
	Size             int64      `json:"size"`
	Kind             string     `json:"kind"`
	Access           AccessType `json:"access"`
	ShareAccess      AccessType `json:"share_access"`
}
type CommandsSearchParameters struct {
	Status *EnumStatus `json:"status,omitempty"`
}

type BrowserCommandBrowseResult struct {
	CurrentItem StorageItem   `json:"current_item"` //Information on current Path
	Children    []StorageItem `json:"children"`
}

//Browse command structure. This is used for request and response
type BrowserCommandList struct {
	Path   string                     `json:"path"`
	Result BrowserCommandBrowseResult `json:"result"`
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

type BrowserUploadFile struct {
	Path string `json:"path"` //Path of the file it will be uploaded to
	Size int64  `json:"size"` //Size of the file. This will be used for computing the progress
}

type EnumShareLinkType string

const (
	EnumShareByKey    EnumShareLinkType = "key"           //EveryBody with the key can access to this sharelink
	EnumRestricted    EnumShareLinkType = "restricted"    // Shared to a limited number of users can access to this link
	EnumAuthenticated EnumShareLinkType = "authenticated" //Only the authenticated users with the key can access to this sharelink
)

type ShareLink struct {
	Name     *string           `json:"name,omitempty"`      //Name used for displaying the share link if multiple share links available.
	Path     *string           `json:"path,omitempty"`      //Can be empty only if ShareLinkKey is provided
	Key      *string           `json:"key,omitempty"`       //Can be empty only for a creation or on a Get
	User     string            `json:"user"`                //This will only be set by server. This is the user that issued the share link
	UserList *[]string         `json:"user_list,omitempty"` //This is only available for EnumRestricted mode
	Type     EnumShareLinkType `json:"type"`
	Access   *AccessType       `json:"access,omitempty"` //What access would people coming with this link have
}

type ShareLinkCreate struct {
	ShareLink ShareLink `json:"share_link"`
}

type ShareLinkUpdate struct {
	ShareLink ShareLink `json:"share_link"`
}

type ShareLinkGet struct {
	Path    string       `json:"path"`
	Results []*ShareLink `json:"results"`
}

type ShareLinkCommand struct {
	Create *ShareLinkCreate `json:"create,omitempty"`
	Update *ShareLinkUpdate `json:"update,omitempty"` //Key field needs to be provided
	Delete *string          `json:"delete,omitempty"` //Just pass the key and we'll be find
	List   *ShareLinkGet    `json:"list,omitempty"`   //List ShareLinks from a path for the current user. This will only return if users is owner
}

type BrowserCommand struct {
	List                 *BrowserCommandList         `json:"list,omitempty"`
	Delete               *BrowserCommandDeleteItem   `json:"delete,omitempty"`
	CreateFolder         *BrowserCommandCreateFolder `json:"create_folder,omitempty"`
	GenerateDownloadLink *BrowserCommandDownloadLink `json:"download_link,omitempty"`
	UploadFile           *BrowserUploadFile          `json:"upload_file,omitempty"`
}

type Command struct {
	Name      EnumAction        `json:"name"`       // Name of action Requested
	CommandId string            `json:"command_id"` // Command Id returned by client when timeout is reached
	State     CommandStatus     `json:"state"`
	Timeout   int               `json:"timeout"`            // Result should be returned before timeout, or client will have to poll using CommandId
	User      *string           `json:"user,omitempty"`     //This is only used internally to know who is actually making the request
	AuthKey   *string           `json:"auth_key,omitempty"` //Used when calling commands on behalf of a sharedlink
	ShareLink *ShareLinkCommand `json:"share_link,omitempty"`
	Browser   *BrowserCommand   `json:"browser,omitempty"`
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
	EnumBrowserUploadFile   EnumAction = "browser.upload_file"
	EnumDebugLongRequest    EnumAction = "debug.long_request"
	EnumShareLinkCreate     EnumAction = "share_link.create"
	EnumShareLinkUpdate     EnumAction = "share_link.update"
	EnumShareLinkDelete     EnumAction = "share_link.delete"
	EnumShareLinkList       EnumAction = "share_link.list"
)

type EnumCommandHandlerStatus int

const (
	EnumCommandHandlerError     EnumCommandHandlerStatus = 0
	EnumCommandHandlerDone      EnumCommandHandlerStatus = 1
	EnumCommandHandlerPostponed EnumCommandHandlerStatus = 2
)

type HttpError struct {
	Err    error
	Status int
}

type CommandContext struct {
	Command *Command
	Account *Account
	Request *http.Request
}

type CommandHandler interface {
	Handle(context *CommandContext, resp chan<- EnumCommandHandlerStatus) *HttpError
	GetUploadPath(context *CommandContext) (path *string, resultFileSize int64, hErr *HttpError)
}
