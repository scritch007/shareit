package dummy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jmcvetta/randutil"

	"github.com/labstack/echo/v4"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/types"
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
	if err = json.Unmarshal(*config, d); err != nil {
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
func (d *DummyAuth) AddRoutes(r *echo.Echo) error {
	r.POST(path.Join(d.config.HtmlPrefix, api.RequestDummyAuthUrl), func(ctx echo.Context) error {
		return d.HandleAuth(ctx)

	})
	r.GET(path.Join(d.config.HtmlPrefix, api.RequestDummyGetChallengeUrl), func(ctx echo.Context) error {
		return d.HandleGetChallenge(ctx)

	})
	r.POST(path.Join(d.config.HtmlPrefix, api.RequestDummyCreateUrl), func(ctx echo.Context) error {
		return d.HandleCreate(ctx)
	})
	return nil
}

func (auth *DummyAuth) HandleAuth(ctx echo.Context) error {
	//method auth, create, validate
	var authCommand api.RequestDummyAuthInput

	if err := ctx.Bind(&authCommand); err != nil {
		ctx.Error(fmt.Errorf("Couldn't parse command %w", err))
		return nil
	}
	splittedElements := strings.Split(authCommand.ChallengeHash, ":")
	if 2 != len(splittedElements) {
		tools.LOG_ERROR.Println("Wrong Input")
		return ctx.String(http.StatusBadRequest, "Wrong command Input")
	}
	account, _, err := auth.config.Db.GetAccount(Name, authCommand.Login)
	if err != nil {
		tools.LOG_ERROR.Println("Unknown user ", authCommand.Login, " requested")

		return ctx.String(http.StatusUnauthorized, fmt.Sprintf("Unknown user %s requested", authCommand.Login))
	}
	refInt, err := strconv.Atoi(authCommand.Ref)
	if err != nil {
		tools.LOG_ERROR.Println("Invalid reference ", authCommand.Ref)
		return ctx.String(http.StatusUnauthorized, fmt.Sprintf("Invalid Reference"))
	}
	challenge, found := auth.challengeMap[refInt]
	if !found {
		tools.LOG_ERROR.Println("Challenge ,", authCommand.Ref, " not found ")
		return ctx.String(http.StatusUnauthorized, fmt.Sprintf("Challenge ,%s not found", authCommand.Ref))

	}

	if challenge.apiChallenge.Challenge != splittedElements[0] {
		tools.LOG_ERROR.Println("Incorrect Challenge Hash ", splittedElements[0], " received, expecting :", challenge.apiChallenge.Challenge)
		return ctx.String(http.StatusUnauthorized, "Incorrect Challenge Hash")

	}
	authSpecific := account.Auths[Name]
	if authSpecific.Blob != splittedElements[1] {
		tools.LOG_ERROR.Println("Incorrect Password ", splittedElements[1], " received, expecting :", authSpecific.Blob)
		return ctx.String(http.StatusUnauthorized, "Incorrect Password")

	}
	challenge.timer.Stop()
	delete(auth.challengeMap, refInt)
	var apiAccount api.Account
	err = types.AccountBackend2Api(account, &apiAccount)
	if err != nil {
		tools.LOG_ERROR.Println("Convertion from backend Account to api Account failed")

		return ctx.String(http.StatusUnauthorized, "Couldn't save session")
	}
	result := api.RequestDummyAuthOutput{AuthenticationHeader: account.Email, MySelf: apiAccount}
	session := types.Session{AuthenticationHeader: result.AuthenticationHeader, UserId: account.Id}
	err = auth.config.Db.StoreSession(&session)
	if err != nil {
		tools.LOG_ERROR.Println("Couldn't save session")

		return ctx.String(http.StatusUnauthorized, "Couldn't save session")
	}

	return ctx.JSON(http.StatusOK, &result)
}

func (auth *DummyAuth) HandleCreate(ctx echo.Context) error {

	var create api.RequestDummyCreateInput

	if err := ctx.Bind(&create); err != nil {
		return ctx.String(http.StatusBadRequest, "Couldn't parse command")
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
	err := auth.config.Db.AddAccount(account)
	if err != nil {
		errMessage := fmt.Sprintf("Couldn't save this account with error %s", err)
		tools.LOG_ERROR.Println(errMessage)

		return ctx.String(http.StatusInternalServerError, errMessage)
	}
	var apiAccount api.Account
	err = types.AccountBackend2Api(account, &apiAccount)
	result := api.RequestDummyAuthOutput{AuthenticationHeader: account.Email, MySelf: apiAccount}

	return ctx.JSON(http.StatusOK, &result)
}
func (auth *DummyAuth) HandleValidateAccount(ctx echo.Context) error {
	return nil
}
func (auth *DummyAuth) HandleGetChallenge(ctx echo.Context) error {

	var chal challenge
	var err error
	chal.apiChallenge.Challenge, err = randutil.AlphaString(20)
	if err != nil {
		tools.LOG_ERROR.Println(fmt.Sprintf("Failed to generate Random string with error: %s", err))
		return ctx.String(http.StatusBadRequest, "Failed to generate Random String")
	}
	challengeId := auth.challengeId
	chal.apiChallenge.Ref = strconv.Itoa(challengeId)
	auth.challengeMap[challengeId] = &chal
	chal.timer = time.AfterFunc(2*time.Second, func() {
		tools.LOG_DEBUG.Println("Removing challenge ", challengeId, ", it has just expired")
		delete(auth.challengeMap, challengeId)
	})

	auth.challengeId += 1

	return ctx.JSON(http.StatusOK, chal.apiChallenge)
}
