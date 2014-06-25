package share_link

import (
	//"encoding/json"
	"github.com/scritch007/shareit/types"
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

	if command.Name == types.EnumShareLinkCreate {

	} else if command.Name == types.EnumShareLinkUpdate {

	} else if command.Name == types.EnumShareLinkDelete {

	} else {
		//Unknown command....
		resp <- false
	}
}
