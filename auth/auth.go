package auth

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit/auth/dummy"
	"github.com/scritch007/shareit/types"
)

//Should be called by authentication mechanism
func NewAuthentication(auth string, config *json.RawMessage, r *mux.Router, globalConfig *types.Configuration) (newAuth types.Authentication, err error) {
	switch auth {
	case dummy.Name:
		newAuth, err = dummy.NewDummyAuth(config, globalConfig)
	default:
		err = errors.New("Unknown authentication method " + auth)
		newAuth = nil
	}
	if nil == err {
		newAuth.AddRoutes(r)
	}
	return newAuth, err
}
