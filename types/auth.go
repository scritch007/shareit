package types

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
)

type SubAuthentication interface {
	Name() string
	AddRoutes(r *echo.Echo) error
}

type Authentication struct {
	Auths  []SubAuthentication
	Config *Configuration
}

func (auth *Authentication) GetAuthenticatedUser(w http.ResponseWriter, r *http.Request) (user *Account, err error) {
	//TODO KeyWord should be changed
	authHeader := r.Header.Get("Authentication")
	user = nil
	err = nil

	if len(authHeader) == 0 {
		authHeader = r.URL.Query().Get("Authentication")
		if len(authHeader) == 0 {
			return nil, nil
		}
	}
	session, err := auth.Config.Db.GetSession(authHeader)
	if nil != err {
		return nil, err
	}
	userAccount, err := auth.Config.Db.GetUserAccount(session.UserId)
	if nil != err {
		return nil, err
	}
	return userAccount, err

}

func (auth *Authentication) LogOut(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authentication")
	if len(authHeader) == 0 {
		authHeader = r.URL.Query().Get("Authentication")
		if len(authHeader) == 0 {
			http.Error(w, "You're not logged in", http.StatusBadRequest)
			return
		}
		return
	}
	err := auth.Config.Db.RemoveSession(authHeader)
	if nil != err {
		http.Error(w, "You're not logged in", http.StatusBadRequest)
	}
}

func (auth *Authentication) GetAvailableAuthentications() []string {
	res := make([]string, len(auth.Auths))
	for i, elem := range auth.Auths {
		res[i] = elem.Name()
	}
	return res
}

func (auth *Authentication) ListUsers(w http.ResponseWriter, r *http.Request) {
	user, err := auth.GetAuthenticatedUser(w, r)
	if nil == user {
		var message string
		if nil != err {
			message = err.Error()
		} else {
			message = "You're not allowed to do this"
		}
		http.Error(w, message, http.StatusUnauthorized)
		return
	}
	search := r.URL.Query().Get("search")
	searchParameters := make(map[string]string)
	if len(search) != 0 {
		searchParameters["login"] = search
		searchParameters["id"] = search
	}
	accounts, _ := auth.Config.Db.ListAccounts(searchParameters)
	resp := make([]api.Account, len(accounts))
	for i, account := range accounts {
		AccountBackend2Api(account, &resp[i])
	}
	//var tempResult []*Account
	b, _ := json.Marshal(resp)
	io.WriteString(w, string(b))
}

func (auth *Authentication) GetInfo(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	shareInfo := api.RequestGetInfoOutput{}
	if len(key) == 0 {
		shareInfo.ShareLink = false
		b, _ := json.Marshal(shareInfo)
		io.WriteString(w, string(b))
		return
	}
	result, err := auth.Config.Db.GetShareLink(key)
	if err != nil {
		shareInfo.ShareLink = false
		b, _ := json.Marshal(shareInfo)
		io.WriteString(w, string(b))
		return
	}
	shareInfo.ShareLink = true

	if result.ShareLink.Type == api.EnumShareByKeyAndPassword {
		shareInfo.PasswordProtected = true
	} else {
		shareInfo.PasswordProtected = false
	}

	shareInfo.NbDownloads = result.ShareLink.NbDownloads
	shareInfo.Type = result.ShareLink.Type
	shareInfo.Access = *result.ShareLink.Access
	shareInfo.ShareAccess = *result.ShareLink.Access
	b, _ := json.Marshal(shareInfo)
	io.WriteString(w, string(b))
}
