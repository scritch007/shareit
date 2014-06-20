package dummy

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit/types"
	"io/ioutil"
	"net/http"
)

type DummyAuth struct {
	AutoValidateAccount bool   `json:"autovalidate"`
	GmailLogin          string `json:"gmail_login"`
	GmailPassword       string `json:"gmail_password"`
	config              *types.Configuration
}

func NewDummyAuth(config *json.RawMessage, globalConfig *types.Configuration) (d *DummyAuth, err error) {
	d = new(DummyAuth)
	if err = json.Unmarshal(*config, d); nil != err {
		return nil, err
	}
	d.config = globalConfig
	return d, nil
}

const (
	Name string = "DummyAuth"
)

func (d *DummyAuth) Name() string {
	return Name
}
func (d *DummyAuth) AddRoutes(r *mux.Router) error {
	r.HandleFunc("/auths/dummy/{method}", d.Handle).Methods("POST")
	return nil
}

type CreateCommand struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type GetChallenge struct {
	Login     *string `json:"login,omitempty"`
	Challenge string  `json:"challenge"`
	Ref       string  `json:"ref"`
}

type Auth struct {
	ChallengeHash string `json:"challenge_hash"`
	Ref           string `json:"ref"`
}

func (auth *DummyAuth) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	method := vars["method"]
	//method auth, create, validate
	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		fmt.Println("1 Failed with error code " + err.Error())
		return
	}

	switch method {
	case "create":
		var create CreateCommand
		err = json.Unmarshal(input, &create)
		if nil != err {
			types.LOG_ERROR.Println(fmt.Sprintf("Couldn't parse command %s\n with error: %s", input, err))
			http.Error(w, "Couldn't parse command", http.StatusBadRequest)
			return
		}
		account := new(types.Account)
		account.Login = create.Login
		account.Email = create.Email
		account.AuthType = Name
		//TODO This should be the sha1 from the password
		account.Blob = create.Password
		err = auth.config.Db.AddAccount(account)
		if nil != err {
			errMessage := fmt.Sprintf("Couldn't save this account with error %s", err)
			types.LOG_ERROR.Println(errMessage)
			http.Error(w, errMessage, http.StatusInternalServerError)
			return
		}
	case "auth":
	case "validate":
	case "get_challenge":
	case "log_out":
	default:
	}
}
