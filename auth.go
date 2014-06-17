package shareit

import (
	"fmt"
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
	"io/ioutil"
)

type Authentication struct{

}

func NewAuthentication(config *Configuration)(auth *Authentication){
	auth = new(Authentication)
	return auth
}

type createCommand struct{
	Login string `json:"login"`
	Password string `json:"password"`
	Email string `json:"email"`
}

type getChallenge struct {
	Login *string `json:"login,omitempty"`
	Challenge string `json:"challenge"`
	Ref string `json:"ref"`
}

type auth struct{
	ChallengeHash string `json:"challenge_hash"`
	Ref string `json:"ref"`
}

func (auth *Authentication)Handle(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	method := vars["method"]
	//method auth, create, validate
	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		fmt.Println("1 Failed with error code " + err.Error())
		return
	}

	if "create" == method{
		var create createCommand
		err = json.Unmarshal(input, create)
		if nil != err {
			fmt.Println("Couldn't parse command")
			//TODO write response error
			return
		}
	}else if("auth" == method){

	}else if("validate" == method){

	}else if("get_challenge" == method){
		
	}
}