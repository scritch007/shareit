package share_link

import (
	//"encoding/json"
	"github.com/jmcvetta/randutil"
	"github.com/scritch007/shareit/types"
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

func (s *ShareLinkHandler) Handle(command *types.Command, resp chan<- bool) {
	if nil == command.User {
		//only users can play with the share links
		command.State.ErrorCode = types.ERROR_NOT_ALLOWED
		resp <- false
		return
	}

	if command.Name == types.EnumShareLinkCreate {
		if nil == command.ShareLink.Create {
			types.LOG_DEBUG.Println("Missing input configuration")
			command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
			resp <- false
			return
		}
		//TODO check that the path is provided and that it is a valid path.
		if nil == command.ShareLink.Create.ShareLink.Path {
			types.LOG_DEBUG.Println("Missing path parameter")
			command.State.ErrorCode = types.ERROR_MISSING_PARAMETERS
			resp <- false
			return
		}
		item_path := path.Join(s.config.RootPrefix, *command.ShareLink.Create.ShareLink.Path)
		types.LOG_DEBUG.Println("Creating sharelink for " + item_path)

		fileInfo, err := os.Lstat(item_path)
		if nil != err {
			types.LOG_ERROR.Println("Couldn't access to path ", item_path)
			command.State.ErrorCode = types.ERROR_INVALID_PATH
			resp <- false
			return
		}
		if !fileInfo.IsDir() {
			types.LOG_ERROR.Println(item_path, " is not a directory...")
			command.State.ErrorCode = types.ERROR_INVALID_PATH
			resp <- false
			return
		}
		key, err := randutil.AlphaString(20)
		command.ShareLink.Create.ShareLink.Key = &key
		command.ShareLink.Create.ShareLink.User = *command.User
		*command.ShareLink.Create.ShareLink.Path = path.Clean(*command.ShareLink.Create.ShareLink.Path)
		err = s.config.Db.SaveShareLink(&command.ShareLink.Create.ShareLink)
		if nil != err {
			resp <- false
			types.LOG_ERROR.Println("Failed to save the Sharelink")
			command.State.ErrorCode = types.ERROR_SAVING
			return
		}
		resp <- true

	} else if command.Name == types.EnumShareLinkUpdate {

	} else if command.Name == types.EnumShareLinkDelete {

	} else if command.Name == types.EnumShareLinkGet {
		shareLink, err := s.config.Db.GetShareLinkFromPath(command.ShareLink.Get.Path, *command.User)
		if nil != err {
			resp <- false
		}
		command.ShareLink.Get.Result = shareLink
		resp <- true
	} else {
		//Unknown command....
		resp <- false
	}
}
