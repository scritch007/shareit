package auth

import (
	"github.com/gorilla/mux"
	"github.com/scritch007/shareit/auth/dummy"
	"encoding/json"
	"errors"
)

type Authentication interface {
	Name() string
	AddRoutes(r *mux.Router) error
}


//Should be called by authentication mechanism
func NewAuthentication(auth string, config json.RawMessage, r *mux.Router) (newAuth Authentication, err error){
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