package share_link

import (
	//"encoding/json"
	"errors"
	"github.com/jmcvetta/randutil"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/shareit/types"
	"github.com/scritch007/go-tools"
	"net/http"
	"os"
	"path"
)

var COMMAND_PREFIX = "share_link"

type ShareLinkHandler struct {
	config *types.Configuration
}

func NewShareLinkHandler(config *types.Configuration) (s *ShareLinkHandler) {
	s = new(ShareLinkHandler)
	s.config = config
	return s
}

func (s *ShareLinkHandler) create(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command.ApiCommand
	if nil == command.ShareLink.Create {
		tools.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = api.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	//TODO check that the path is provided and that it is a valid path.
	if nil == command.ShareLink.Create.Input.ShareLink.Path {
		tools.LOG_DEBUG.Println("Missing path parameter")
		command.State.ErrorCode = api.ERROR_MISSING_PARAMETERS
		resp <- types.EnumCommandHandlerError
		return
	}
	item_path := path.Join(s.config.RootPrefix, *command.ShareLink.Create.Input.ShareLink.Path)
	tools.LOG_DEBUG.Println("Creating sharelink for " + item_path)

	_, err := os.Lstat(item_path)
	if nil != err {
		tools.LOG_ERROR.Println("Couldn't access to path ", item_path)
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}
	//Remove this part allow sharing a single file
	//if !fileInfo.IsDir() {
	//	tools.LOG_ERROR.Println(item_path, " is not a directory...")
	//	command.State.ErrorCode = &api.ERROR_INVALID_PATH
	//	resp <- types.EnumCommandHandlerError
	//	return
	//}
	key, err := randutil.AlphaString(20)
	command.ShareLink.Create.Input.ShareLink.Key = &key

	//*command.ShareLink.Create.Input.ShareLink.Path = path.Clean(*command.ShareLink.Create.Input.ShareLink.Path)
	var sLink types.ShareLink
	sLink.ShareLink = command.ShareLink.Create.Input.ShareLink
	*sLink.ShareLink.Path = path.Clean(*command.ShareLink.Create.Input.ShareLink.Path)
	sLink.User = context.Account.Email
	if nil == command.ShareLink.Create.Input.ShareLink.Name{
		sLink.ShareLink.Name = new(string)
		*sLink.ShareLink.Name = "share_link"
	}
	//sLink.Key = *command.ShareLink.Create.Input.ShareLink.Key
	//sLink.User = command.ShareLink.Create.Input.ShareLink.User
	//TODO check that the access is correct and is not more then the user actually has access to

	//sLink.Users = command.ShareLink.Create.Input.ShareLink.Users
	err = s.config.Db.SaveShareLink(&sLink)
	command.ShareLink.Create.Output.ShareLink = sLink.ShareLink
	if nil != err {
		tools.LOG_ERROR.Println("Failed to save the Sharelink")
		command.State.ErrorCode = api.ERROR_SAVING
		resp <- types.EnumCommandHandlerError
		return
	}
	resp <- types.EnumCommandHandlerDone
}

func (s *ShareLinkHandler) get(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command.ApiCommand
	shareLinks, err := s.config.Db.GetShareLinksFromPath(command.ShareLink.List.Input.Path, *context.Command.User)
	if nil != err {
		resp <- types.EnumCommandHandlerError
	}

	//TODO sharelinks... iteration
	command.ShareLink.List.Output.ShareLinks = make([]api.ShareLink, len(shareLinks))
	for i, sLink := range shareLinks {
		command.ShareLink.List.Output.ShareLinks[i] = sLink.ShareLink
	}
	resp <- types.EnumCommandHandlerDone
}

func (s *ShareLinkHandler) Handle(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) *types.HttpError {
	command := context.Command
	if nil == command.User {
		//only users can play with the share links
		return &types.HttpError{errors.New("Method requires"), http.StatusUnauthorized}
	}
	if command.ApiCommand.Name == api.EnumShareLinkCreate {
		go s.create(context, resp)
	} else if command.ApiCommand.Name == api.EnumShareLinkUpdate {

	} else if command.ApiCommand.Name == api.EnumShareLinkDelete {

	} else if command.ApiCommand.Name == api.EnumShareLinkList {
		go s.get(context, resp)
	} else {
		return &types.HttpError{errors.New("Unknown share_link command"), http.StatusBadRequest}
	}
	return nil
}

func (s *ShareLinkHandler) GetUploadPath(context *types.CommandContext) (*string, int64, *types.HttpError) {
	return nil, 0, &types.HttpError{errors.New("Not Allowed"), http.StatusBadRequest}
}
