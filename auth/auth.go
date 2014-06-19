package auth

import (
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit/auth/dummy"
	"github.com/scritch007/shareit/types"
	"encoding/json"
	"errors"
)

//Should be called by authentication mechanism
func NewAuthentication(auth string, config *json.RawMessage, r *mux.Router) (newAuth types.Authentication, err error){
	switch auth{
	case dummy.Name:
		newAuth, err = dummy.NewDummyAuth(config)
	default:
		err = errors.New("Unknown authentication method " + auth)
		newAuth = nil
	}
	if nil == err{
		newAuth.AddRoutes(r)
	}
	return newAuth, err
}