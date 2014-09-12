package types

import (
	"github.com/scritch007/ShareMinatorApiGenerator/api"
)

func AccountApi2Backend(apiAccount *api.Account, bAccount *Account) error {
	bAccount.Login = apiAccount.Login
	bAccount.Id = apiAccount.Id
	bAccount.IsAdmin = apiAccount.IsAdmin
	bAccount.Email = apiAccount.Email
	return nil
}

func AccountBackend2Api(bAccount *Account, apiAccount *api.Account) error {
	apiAccount.Login = bAccount.Login
	apiAccount.Id = bAccount.Id
	apiAccount.IsAdmin = bAccount.IsAdmin
	apiAccount.Email = bAccount.Email
	return nil
}
