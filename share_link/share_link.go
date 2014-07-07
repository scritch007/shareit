package share_link

import (
	//"encoding/json"
	"errors"
	"github.com/jmcvetta/randutil"
	"github.com/scritch007/shareit/types"
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
	command := context.Command
	if nil == command.ShareLink.Create {
		types.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	//TODO check that the path is provided and that it is a valid path.
	if nil == command.ShareLink.Create.ShareLink.Path {
		types.LOG_DEBUG.Println("Missing path parameter")
		command.State.ErrorCode = types.ERROR_MISSING_PARAMETERS
		resp <- types.EnumCommandHandlerError
		return
	}
	item_path := path.Join(s.config.RootPrefix, *command.ShareLink.Create.ShareLink.Path)
	types.LOG_DEBUG.Println("Creating sharelink for " + item_path)

	fileInfo, err := os.Lstat(item_path)
	if nil != err {
		types.LOG_ERROR.Println("Couldn't access to path ", item_path)
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}
	if !fileInfo.IsDir() {
		types.LOG_ERROR.Println(item_path, " is not a directory...")
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}
	key, err := randutil.AlphaString(20)
	command.ShareLink.Create.ShareLink.Key = &key
	command.ShareLink.Create.ShareLink.User = *command.User
	*command.ShareLink.Create.ShareLink.Path = path.Clean(*command.ShareLink.Create.ShareLink.Path)
	err = s.config.Db.SaveShareLink(&command.ShareLink.Create.ShareLink)
	if nil != err {
		types.LOG_ERROR.Println("Failed to save the Sharelink")
		command.State.ErrorCode = types.ERROR_SAVING
		resp <- types.EnumCommandHandlerError
		return
	}
	resp <- types.EnumCommandHandlerDone
}

func (s *ShareLinkHandler) get(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command
	shareLink, err := s.config.Db.GetShareLinkFromPath(command.ShareLink.Get.Path, *command.User)
	if nil != err {
		resp <- types.EnumCommandHandlerError
	}
	command.ShareLink.Get.Result = shareLink
	resp <- types.EnumCommandHandlerDone
}

func (s *ShareLinkHandler) Handle(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) *types.HttpError {
	command := context.Command
	if nil == command.User {
		//only users can play with the share links
		return &types.HttpError{errors.New("Method requires"), http.StatusUnauthorized}
	}
	if command.Name == types.EnumShareLinkCreate {
		go s.create(context, resp)
	} else if command.Name == types.EnumShareLinkUpdate {

	} else if command.Name == types.EnumShareLinkDelete {

	} else if command.Name == types.EnumShareLinkGet {
		go s.get(context, resp)
	} else {
		return &types.HttpError{errors.New("Unknown share_link command"), http.StatusBadRequest}
	}
	return nil
}

func (s *ShareLinkHandler) GetUploadPath(context *types.CommandContext) (*string, int64, *types.HttpError) {
	return nil, 0, &types.HttpError{errors.New("Not Allowed"), http.StatusBadRequest}
}
