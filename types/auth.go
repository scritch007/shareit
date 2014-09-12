package types

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"io"
	"net/http"
)

type SubAuthentication interface {
	Name() string
	AddRoutes(r *mux.Router) error
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
	if 0 != len(search) {
		searchParameters["login"] = search
		searchParameters["id"] = search
	}
	accounts, err := auth.Config.Db.ListAccounts(searchParameters)
	resp := make([]api.Account, len(accounts))
	for i, account := range accounts {
		resp[i].Login = account.ApiAccount.Login
		resp[i].Id = account.ApiAccount.Id
		resp[i].IsAdmin = account.ApiAccount.IsAdmin
		resp[i].Email = account.ApiAccount.Email
	}
	//var tempResult []*Account
	b, _ := json.Marshal(resp)
	io.WriteString(w, string(b))
}
