package browse

import (
	"errors"
	"github.com/scritch007/shareit/tools"
	"github.com/scritch007/shareit/types"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

type BrowseHandler struct {
	config *types.Configuration
}

func NewBrowseHandler(config *types.Configuration) (handler *BrowseHandler) {
	handler = &BrowseHandler{config: config}
	return handler
}

func (b *BrowseHandler) Handle(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) *types.HttpError {
	command := context.Command
	if nil == command.Browser {
		return &types.HttpError{Err: errors.New("Missing browse command body"), Status: http.StatusBadRequest}
	}
	if command.Name == types.EnumBrowserBrowse {
		go b.browseCommand(context, resp)
	} else if command.Name == types.EnumBrowserCreateFolder {
		go b.createFolderCommand(context, resp)
	} else if command.Name == types.EnumBrowserDeleteItem {
		go b.deleteItemCommand(context, resp)
	} else if command.Name == types.EnumBrowserDownloadLink {
		go b.downloadLink(context, resp)
	} else if command.Name == types.EnumBrowserUploadFile {
		go b.uploadFile(context, resp)
	} else {
		return &types.HttpError{Err: errors.New("Unknown Browse command"), Status: http.StatusBadRequest}
	}
	return nil
}

func (b *BrowseHandler) downloadLink(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command
	if nil == command.Browser.GenerateDownloadLink {
		types.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	file_path := path.Join(b.config.RootPrefix, command.Browser.GenerateDownloadLink.Path)
	result := tools.ComputeHmac256(file_path, b.config.PrivateKey)
	dLink := types.DownloadLink{Link: result, Path: command.Browser.GenerateDownloadLink.Path}
	b.config.Db.AddDownloadLink(&dLink)
	command.Browser.GenerateDownloadLink.Result.Link = url.QueryEscape(result)
	command.Browser.GenerateDownloadLink.Result.Path = command.Browser.GenerateDownloadLink.Path
	resp <- types.EnumCommandHandlerDone
}

//Handle removal of an item
func (b *BrowseHandler) deleteItemCommand(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command
	if nil == command.Browser.Delete {
		types.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	item_path, fileInfo := b.checkItemPath(&command.Browser.Delete.Path)
	if nil == item_path {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}
	if fileInfo.IsDir() {
		types.LOG_DEBUG.Println("Item is a directory")
		//We are going to make something nice with a progress
		fileList, err := ioutil.ReadDir(*item_path)
		if nil != err {
			types.LOG_DEBUG.Println("Couldn't list directory")
			command.State.ErrorCode = types.ERROR_FILE_SYSTEM
			resp <- types.EnumCommandHandlerError
			return
		}
		nbElements := len(fileList)
		success := types.EnumCommandHandlerDone
		for i, element := range fileList {
			element_path := path.Join(*item_path, element.Name())
			types.LOG_DEBUG.Println("Trying to remove " + element_path)
			err = os.RemoveAll(element_path)
			if nil != err {
				success = types.EnumCommandHandlerError
				command.State.ErrorCode = types.ERROR_FILE_SYSTEM
			}
			command.State.Progress = i * 100 / nbElements
		}
		if nil != os.RemoveAll(*item_path) {
			success = types.EnumCommandHandlerError
		}
		resp <- success
	} else {
		err := os.Remove(*item_path)
		if nil == err {
			resp <- types.EnumCommandHandlerDone
		} else {
			resp <- types.EnumCommandHandlerError
		}
	}

}

//Extend the path with the RootPrefix and check if it exists.
func (b *BrowseHandler) checkItemPath(inPath *string) (*string, os.FileInfo) {
	item_path := path.Join(b.config.RootPrefix, *inPath)
	fileInfo, err := os.Lstat(item_path)
	if nil != err {
		if os.IsNotExist(err) {
			return &item_path, nil
		}
		return nil, nil
	}
	return &item_path, fileInfo
}

//Handle the creation of a folder
func (b *BrowseHandler) createFolderCommand(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command
	if nil == command.Browser.CreateFolder {
		types.LOG_ERROR.Println("Missing configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	error := os.Mkdir(path.Join(b.config.RootPrefix, command.Browser.CreateFolder.Path), os.ModePerm)
	if nil != error {
		resp <- types.EnumCommandHandlerError
	} else {
		resp <- types.EnumCommandHandlerDone
	}
}

func (b *BrowseHandler) uploadFile(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command
	if nil == command.Browser.UploadFile {
		types.LOG_ERROR.Println("Missing configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	item_path, fileInfo := b.checkItemPath(&command.Browser.UploadFile.Path)
	if nil == item_path {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	if nil != fileInfo {
		types.LOG_DEBUG.Println("Truncating file")
		os.Truncate(*item_path, command.Browser.UploadFile.Size)
		return
	}
	resp <- types.EnumCommandHandlerPostponed
}

//Handle the browsing of a folder
func (b *BrowseHandler) browseCommand(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command
	if nil == command.Browser.List {
		types.LOG_ERROR.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	asUser := true
	if nil != context.Account && context.Account.IsAdmin{
		//TODO checkt he asUser request header
		asUser = false
	}

	//First check if we have a Key. If we do then we'll chroot the browse command...
	chroot := ""
	access := types.READ // Default access type
	isRoot := "/" == command.Browser.List.Path
	if nil != command.AuthKey {
		share_link, err := b.config.Db.GetShareLink(*command.AuthKey)
		if nil != err {
			command.State.ErrorCode = types.ERROR_INVALID_PARAMETERS
			resp <- types.EnumCommandHandlerError
			return
		}
		chroot = *share_link.Path
		//TODO add some check depending on the type of share_link...
	}else{

		if !isRoot{
			//Check if user has access to this path
			access, err := b.config.Db.GetAccess(command.User, command.Browser.List.Path)
			if nil != err {
				types.LOG_ERROR.Println("Couldn't get access " + err.Error())
				command.State.ErrorCode = types.ERROR_INVALID_PATH
				resp <-types.EnumCommandHandlerError
				return
			}
			if types.NONE == access && asUser{
				command.State.ErrorCode = types.ERROR_NOT_ALLOWED
				resp <- types.EnumCommandHandlerError
				return
			}
		}
	}
	realPath := path.Join(b.config.RootPrefix, chroot, command.Browser.List.Path)
	types.LOG_DEBUG.Println("Browsing path ", realPath)
	fileList, err := ioutil.ReadDir(realPath)
	if nil != err {
		types.LOG_ERROR.Println("Failed to read path with error code " + err.Error())
		resp <- types.EnumCommandHandlerError
	}
	var result = make([]types.StorageItem, len(fileList))
	counter := 0
	for _, file := range fileList {
		s := types.StorageItem{Name: file.Name(), IsDir: file.IsDir(), ModificationDate: file.ModTime().Unix()}
		if !file.IsDir() {
			s.Size = file.Size()
			s.Kind = filepath.Ext(file.Name())
		} else {
			s.Kind = "folder"
		}
		if isRoot{
			access, err = b.config.Db.GetAccess(command.User, path.Join("/", s.Name))
			if nil != err{
				continue
			}else if (types.NONE == access && asUser){
				continue
			}
			s.Access = access
		}else{
			s.Access = access
		}
		result[counter] = s
		counter++
	}
	command.Browser.List.Results = result[:counter]
	time.Sleep(2)
	resp <- types.EnumCommandHandlerError
}

func (b *BrowseHandler) GetUploadPath(context *types.CommandContext) (path *string, size int64, hErr *types.HttpError) {
	command := context.Command
	if types.EnumBrowserUploadFile != command.Name {
		return nil, 0, &types.HttpError{errors.New("Not Allowed for this command type"), http.StatusBadRequest}
	}
	item_path, _ := b.checkItemPath(&command.Browser.UploadFile.Path)
	if nil == item_path {
		return nil, 0, &types.HttpError{errors.New("Invalid parameter"), http.StatusBadRequest}
	}
	return item_path, command.Browser.UploadFile.Size, nil
}
