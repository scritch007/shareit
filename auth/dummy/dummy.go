package dummy

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmcvetta/randutil"
	"github.com/scritch007/shareit/types"
	"github.com/scritch007/go-tools"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

type DummyAuth struct {
	AutoValidateAccount bool   `json:"autovalidate"`
	GmailLogin          string `json:"gmail_login"`
	GmailPassword       string `json:"gmail_password"`
	config              *types.Configuration
	challengeId         int
	challengeMap        map[int]*GetChallenge
}

func NewDummyAuth(config *json.RawMessage, globalConfig *types.Configuration) (d *DummyAuth, err error) {
	d = new(DummyAuth)
	if err = json.Unmarshal(*config, d); nil != err {
		return nil, err
	}
	d.config = globalConfig
	d.challengeId = 0
	d.challengeMap = make(map[int]*GetChallenge)
	return d, nil
}

const (
	Name string = "DummyAuth"
)

func (d *DummyAuth) Name() string {
	return Name
}
func (d *DummyAuth) AddRoutes(r *mux.Router) error {
	r.HandleFunc(path.Join(d.config.HtmlPrefix, "/auths/", Name, "/{method}"), d.Handle).Methods("POST", "GET")
	return nil
}

type CreateCommand struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type GetChallenge struct {
	Challenge string      `json:"challenge"`
	Ref       string      `json:"ref"`
	timer     *time.Timer // Timer used to invalidate token if it takes too long before being used
}

type Auth struct {
	Login         string `json:"login,omitempty"` //Login can be Login or Email
	ChallengeHash string `json:"challenge_hash"`
	Ref           string `json:"ref"`
}

type AuthResult struct {
	AuthenticationHeader string `json:"authentication_header"`
	IsAdmin              bool   `json:"is_admin"`
	Email                string `json:"email"`
	Login                string `json:"login"`
}

func (auth *DummyAuth) Handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	method := vars["method"]
	//method auth, create, validate
	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		tools.LOG_ERROR.Println("1 Failed with error code " + err.Error())
		return
	}

	switch method {
	case "auth":
		var authCommand Auth
		err = json.Unmarshal(input, &authCommand)
		if nil != err {
			tools.LOG_ERROR.Println(fmt.Sprintf("Couldn't parse command %s\n with error: %s", input, err))
			http.Error(w, "Couldn't parse command", http.StatusBadRequest)
			return
		}
		splittedElements := strings.Split(authCommand.ChallengeHash, ":")
		if 2 != len(splittedElements) {
			tools.LOG_ERROR.Println("Wrong Input")
			http.Error(w, "Wrong command Input", http.StatusBadRequest)
			return
		}
		account, _, err := auth.config.Db.GetAccount(Name, authCommand.Login)
		if nil != err {
			tools.LOG_ERROR.Println("Unknown user ", authCommand.Login, " requested")
			http.Error(w, fmt.Sprintf("Unknown user %s requested", authCommand.Login), http.StatusUnauthorized)
			return
		}
		refInt, err := strconv.Atoi(authCommand.Ref)
		if nil != err {
			tools.LOG_ERROR.Println("Invalid reference ", authCommand.Ref)
			http.Error(w, fmt.Sprintf("Invalid Reference"), http.StatusUnauthorized)
			return
		}
		challenge, found := auth.challengeMap[refInt]
		if !found {
			tools.LOG_ERROR.Println("Challenge ,", authCommand.Ref, " not found ")
			http.Error(w, fmt.Sprintf("Challenge ,%s not found", authCommand.Ref), http.StatusUnauthorized)
			return
		}

		if challenge.Challenge != splittedElements[0] {
			tools.LOG_ERROR.Println("Incorrect Challenge Hash ", splittedElements[0], " received, expecting :", challenge.Challenge)
			http.Error(w, "Incorrect Challenge Hash", http.StatusUnauthorized)
			return
		}
		authSpecific := account.Auths[Name]
		if authSpecific.Blob != splittedElements[1] {
			tools.LOG_ERROR.Println("Incorrect Password ", splittedElements[1], " received, expecting :", authSpecific.Blob)
			http.Error(w, "Incorrect Password", http.StatusUnauthorized)
			return
		}
		challenge.timer.Stop()
		delete(auth.challengeMap, refInt)
		result := AuthResult{AuthenticationHeader: account.Email, IsAdmin: account.IsAdmin, Email: account.Email, Login: account.Login}
		session := types.Session{AuthenticationHeader: result.AuthenticationHeader, UserId: account.Id}
		err = auth.config.Db.StoreSession(&session)
		if nil != err {
			tools.LOG_ERROR.Println("Couldn't save session")
			http.Error(w, "Couldn't save session", http.StatusUnauthorized)
			return
		}
		b, _ := json.Marshal(result)
		io.WriteString(w, string(b))

	case "create":
		var create CreateCommand
		err = json.Unmarshal(input, &create)
		if nil != err {
			tools.LOG_ERROR.Println(fmt.Sprintf("Couldn't parse command %s\n with error: %s", input, err))
			http.Error(w, "Couldn't parse command", http.StatusBadRequest)
			return
		}
		account := new(types.Account)
		account.Auths = make(map[string]types.AccountSpecificAuth)
		account.Login = create.Login
		account.Email = create.Email
		account.IsAdmin = false
		authSpecific := types.AccountSpecificAuth{AuthType: Name, Blob: create.Password}
		account.Auths[Name] = authSpecific
		//TODO This should be the sha1 from the password
		err = auth.config.Db.AddAccount(account)
		if nil != err {
			errMessage := fmt.Sprintf("Couldn't save this account with error %s", err)
			tools.LOG_ERROR.Println(errMessage)
			http.Error(w, errMessage, http.StatusInternalServerError)
			return
		}
		result := AuthResult{AuthenticationHeader: account.Email, IsAdmin: account.IsAdmin, Email: account.Email, Login: account.Login}
		b, _ := json.Marshal(result)
		io.WriteString(w, string(b))
	case "validate":
		//TODO
	case "get_challenge":
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		var challenge GetChallenge
		challenge.Challenge, err = randutil.AlphaString(20)
		if nil != err {
			tools.LOG_ERROR.Println(fmt.Sprintf("Failed to generate Random string with error: %s", err))
			http.Error(w, "Failed to generate Random String", http.StatusBadRequest)
		}
		challengeId := auth.challengeId
		challenge.Ref = strconv.Itoa(challengeId)
		auth.challengeMap[challengeId] = &challenge
		challenge.timer = time.AfterFunc(2*time.Second, func() {
			tools.LOG_DEBUG.Println("Removing challenge ", challengeId, ", it has just expired")
			delete(auth.challengeMap, challengeId)
		})

		auth.challengeId += 1
		b, _ := json.Marshal(challenge)
		io.WriteString(w, string(b))
	default:
	}
}
