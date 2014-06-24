package auth

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit/auth/dummy"
	"github.com/scritch007/shareit/types"
)

type configSubStruct struct {
	Type   string          `json:"type"`
	Config *json.RawMessage `json:"config"`
}

//Should be called by authentication mechanism
func NewAuthentication(config *json.RawMessage, r *mux.Router, globalConfig *types.Configuration) (newAuth *types.Authentication, err error) {
	var authConfigs []configSubStruct
	err = json.Unmarshal(*config, &authConfigs)
	newAuth = new(types.Authentication)
	newAuth.Config = globalConfig
	newAuth.Auths = make([]types.SubAuthentication, len(authConfigs))
	var newSubAuth types.SubAuthentication
	for i, elem := range authConfigs {
		switch elem.Type {
		case dummy.Name:
			newSubAuth, err = dummy.NewDummyAuth(elem.Config, globalConfig)
		default:
			err = errors.New("Unknown authentication method " + elem.Type)
			newAuth = nil
		}
		if nil != err{
			return nil, err
		}
		newSubAuth.AddRoutes(r)
		newAuth.Auths[i] = newSubAuth
	}


	return newAuth, err
}

