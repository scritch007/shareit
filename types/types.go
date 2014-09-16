package types

import (
	"encoding/json"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"net/http"
	//  "fmt"
)

type DatabaseInterface interface {
	Name() string
	ListCommands(user *string, offset int, limit int, search_parameters *api.CommandsSearchParameters) ([]*Command, int, error) //set limit to -1 for retrieving all the elements, search_parameters will be used to filter response
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

	GetAccess(user *string, path string) (api.AccessType, error)      //If user is nil then it's public access
	SetAccess(user *string, path string, access api.AccessType) error // This can only be called by a admin...
	ClearAccesses() error                                             //Should only be called when configuration is done by files

	StoreSession(session *Session) (err error)
	GetSession(ref string) (session *Session, err error)
	RemoveSession(ref string) (err error)

	//ShareLink
	SaveShareLink(shareLink *ShareLink) (err error)
	UpdateShareLink(shareLink *ShareLink) (err error)
	GetShareLink(key string) (shareLink *ShareLink, err error)
	ListShareLinks(user string) (shareLinks []*ShareLink, err error) //List the sharelink user created
	GetShareLinksFromPath(path string, user string) (shareLinks []*ShareLink, err error)
	RemoveShareLink(key string) (err error)
}

type AccountSpecificAuth struct {
	AuthType string `json:"auth_type" bson:"auth_type" sql:"type:varchar;"`
	Blob     string `json:"specific" bson:"specific" sql:"type:varchar;"`
}
type Account struct {
	Login   string `json:"login" bson:"login" sql:"type:varchar;"`
	Email   string `json:"email" bson:"email" sql:"type:varchar;"`
	Id      string `json:"id"  bson:"id" sql:"type:varchar;"`             //This id should be unique depending on the DbType
	IsAdmin bool   `json:"is_admin"  bson:"is_admin" sql:"type:boolean;"` //Can only be changed by the admin
	//Add some other infos here for all the specific stuffs
	Auths map[string]AccountSpecificAuth `sql:"-"`
}

type Session struct {
	AuthenticationHeader string `json:"authentication_header" bson:"authentication_header" sql:"type:varchar;"`
	UserId               string `json:"user_id" bson:"user_id" sql:"type:varchar;"`
}

//Configuration Structure
type Configuration struct {
	RootPrefix      string
	PrivateKey      string
	StaticPath      string
	HtmlPrefix      string
	WebPort         string
	AllowRootWrite  bool
	Db              DatabaseInterface
	Auth            *Authentication
	UploadChunkSize int64
}

type DownloadLink struct {
	Link     string  `json:"link" bson:"link" sql:"type:varchar;"`
	Path     string  `json:"path" bson:"path" sql:"type:varchar;"`
	RealPath *string `json:"real_path,omitempty" bson:"real_path,omitempty" sql:"type:varchar;"` //Will only be used when storing in the DB
}

func (d *DownloadLink) String() string {
	b, err := json.Marshal(d)
	if nil != err {
		return "Couldn't serialize"
	}
	return string(b)
}

type ShareLink struct {
	User      string `json:"user" bson:"user"` //This will only be set by server. This is the user that issued the share link
	Id        string `json:"id" bson:"id"`     //shortcut for key
	ShareLink api.ShareLink
}

type Command struct {
	ApiCommand *api.Command `sql:"-"`
	User       *string      `json:"user,omitempty" bson:"user,omitempty" sql:"type:varchar;"` //This is only used internally to know who is actually making the request
	CommandId  string       `json:"command_id" bson:"command_id" sql:"type:varchar;"`
}

func (c *Command) String() string {
	b, err := json.Marshal(c)
	if nil != err {
		return "Couldn't serialize"
	}
	return string(b)
}

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
