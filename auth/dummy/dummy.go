package dummy

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jmcvetta/randutil"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/types"
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
	challengeMap        map[int]*challenge
}

type challenge struct {
	apiChallenge api.RequestDummyGetChallengeOutput
	timer        *time.Timer
}

func NewDummyAuth(config *json.RawMessage, globalConfig *types.Configuration) (d *DummyAuth, err error) {
	d = new(DummyAuth)
	if err = json.Unmarshal(*config, d); nil != err {
		return nil, err
	}
	d.config = globalConfig
	d.challengeId = 0
	d.challengeMap = make(map[int]*challenge)
	return d, nil
}

const (
	Name string = "DummyAuth"
)

func (d *DummyAuth) Name() string {
	return Name
}
func (d *DummyAuth) AddRoutes(r *mux.Router) error {
	r.HandleFunc(path.Join(d.config.HtmlPrefix, api.RequestDummyAuthUrl), d.HandleAuth).Methods("POST")
	r.HandleFunc(path.Join(d.config.HtmlPrefix, api.RequestDummyGetChallengeUrl), d.HandleGetChallenge).Methods("GET")
	r.HandleFunc(path.Join(d.config.HtmlPrefix, api.RequestDummyCreateUrl), d.HandleCreate).Methods("POST")
	return nil
}

func (auth *DummyAuth) HandleAuth(w http.ResponseWriter, r *http.Request) {
	//method auth, create, validate
	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		tools.LOG_ERROR.Println("1 Failed with error code " + err.Error())
		return
	}
	var authCommand api.RequestDummyAuthInput
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

	if challenge.apiChallenge.Challenge != splittedElements[0] {
		tools.LOG_ERROR.Println("Incorrect Challenge Hash ", splittedElements[0], " received, expecting :", challenge.apiChallenge.Challenge)
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
	var apiAccount api.Account
	err = types.AccountBackend2Api(account, &apiAccount)
	if nil != err {
		tools.LOG_ERROR.Println("Convertion from backend Account to api Account failed")
		http.Error(w, "Couldn't save session", http.StatusUnauthorized)
		return
	}
	result := api.RequestDummyAuthOutput{AuthenticationHeader: account.Email, MySelf: apiAccount}
	session := types.Session{AuthenticationHeader: result.AuthenticationHeader, UserId: account.Id}
	err = auth.config.Db.StoreSession(&session)
	if nil != err {
		tools.LOG_ERROR.Println("Couldn't save session")
		http.Error(w, "Couldn't save session", http.StatusUnauthorized)
		return
	}
	b, _ := json.Marshal(result)
	io.WriteString(w, string(b))
}
func (auth *DummyAuth) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.NotFound(w, r)
		return
	}
	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		tools.LOG_ERROR.Println("1 Failed with error code " + err.Error())
		return
	}
	var create api.RequestDummyCreateInput
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
	if nil == create.IsAdmin {
		account.IsAdmin = false
	} else {
		account.IsAdmin = *create.IsAdmin
	}

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
	var apiAccount api.Account
	err = types.AccountBackend2Api(account, &apiAccount)
	result := api.RequestDummyAuthOutput{AuthenticationHeader: account.Email, MySelf: apiAccount}
	b, _ := json.Marshal(result)
	io.WriteString(w, string(b))
}
func (auth *DummyAuth) HandleValidateAccount(w http.ResponseWriter, r *http.Request) {
}
func (auth *DummyAuth) HandleGetChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.NotFound(w, r)
		return
	}
	var chal challenge
	var err error
	chal.apiChallenge.Challenge, err = randutil.AlphaString(20)
	if nil != err {
		tools.LOG_ERROR.Println(fmt.Sprintf("Failed to generate Random string with error: %s", err))
		http.Error(w, "Failed to generate Random String", http.StatusBadRequest)
	}
	challengeId := auth.challengeId
	chal.apiChallenge.Ref = strconv.Itoa(challengeId)
	auth.challengeMap[challengeId] = &chal
	chal.timer = time.AfterFunc(2*time.Second, func() {
		tools.LOG_DEBUG.Println("Removing challenge ", challengeId, ", it has just expired")
		delete(auth.challengeMap, challengeId)
	})

	auth.challengeId += 1
	b, _ := json.Marshal(chal.apiChallenge)
	io.WriteString(w, string(b))
}
