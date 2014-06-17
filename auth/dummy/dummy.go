package dummy

import(
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"fmt"
)

type DummyAuth struct {
	AutoValidateAccount bool `json:"autovalidate"`
	GmailLogin string `json:"gmail_login"`
	GmailPassword string `json:"gmail_password"`
}

func NewDummyAuth(config json.RawMessage) (d *DummyAuth, err error){
	d = new(DummyAuth)
	if err = json.Unmarshal(config, d); nil != err{
		return nil, err
	}
	return d, nil
}
const (
	Name string = "DummyAuth"
)
func (d *DummyAuth)Name() string{
	return Name
}
func (d *DummyAuth)AddRoutes(r *mux.Router) error{
	r.HandleFunc("/auths/dummy/{method}", d.Handle).Methods("POST")
	return nil
}

type CreateCommand struct{
	Login string `json:"login"`
	Password string `json:"password"`
	Email string `json:"email"`
}

type GetChallenge struct {
	Login *string `json:"login,omitempty"`
	Challenge string `json:"challenge"`
	Ref string `json:"ref"`
}

type Auth struct{
	ChallengeHash string `json:"challenge_hash"`
	Ref string `json:"ref"`
}

func (auth *DummyAuth)Handle(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	method := vars["method"]
	//method auth, create, validate
	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		fmt.Println("1 Failed with error code " + err.Error())
		return
	}

	switch method{
		case "create":
		var create CreateCommand
		err = json.Unmarshal(input, create)
		if nil != err {
			fmt.Println("Couldn't parse command")
			//TODO write response error
			return
		}
		case "auth":
		case "validate":
		case "get_challenge":
		default:
	}
}