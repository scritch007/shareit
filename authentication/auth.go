package auth

import (
	"github.com/gorilla/mux"
)

type Authentication interface {
	Name() string
	AddRoutes(r *Router) error
	ParseConfig(configPath string) error
}


//Should be called by authentication mechanism
func RegisterAuthenticationPlugin(){

}

func CreateAccount(login , password, email string) error{

}