package browse

import (
	"errors"
	"github.com/scritch007/shareit/auth"
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

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.GenerateDownloadLink.Path, false)
	if !accessPath.Exists {
		types.LOG_DEBUG.Println("Couldn't find this path")
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	file_path := accessPath.RealPath
	result := tools.ComputeHmac256(*file_path, b.config.PrivateKey)
	dLink := types.DownloadLink{Link: result, Path: command.Browser.GenerateDownloadLink.Path, RealPath: file_path}
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
	asUser := nil == context.Account || !context.Account.IsAdmin

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.Delete.Path, asUser)

	if types.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	if types.READ_WRITE != accessPath.Access {
		command.State.ErrorCode = types.ERROR_NOT_ALLOWED
		resp <- types.EnumCommandHandlerError
		return
	}

	item_path := accessPath.RealPath
	if nil == item_path {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	if !accessPath.Exists {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}
	if accessPath.IsDir {
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

//Handle the creation of a folder
func (b *BrowseHandler) createFolderCommand(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command
	if nil == command.Browser.CreateFolder {
		types.LOG_ERROR.Println("Missing configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	asUser := nil == context.Account || !context.Account.IsAdmin
	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.CreateFolder.Path, asUser)

	if types.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	if types.READ_WRITE != accessPath.Access {
		command.State.ErrorCode = types.ERROR_NOT_ALLOWED
		resp <- types.EnumCommandHandlerError
		return
	}

	item_path := accessPath.RealPath
	if nil == item_path {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	if accessPath.Exists {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	error := os.Mkdir(*item_path, os.ModePerm)
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

	asUser := nil == context.Account || !context.Account.IsAdmin

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.UploadFile.Path, asUser)

	if types.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	item_path := accessPath.RealPath
	if nil == item_path {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	//If write access has been removed, then we should respond false!!
	if types.READ_WRITE != accessPath.Access {
		command.State.ErrorCode = types.ERROR_NOT_ALLOWED
		resp <- types.EnumCommandHandlerError
		return
	}

	if accessPath.Exists {
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
	if nil != context.Account && context.Account.IsAdmin {
		//TODO checkt he asUser request header
		asUser = false
	}

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.List.Path, asUser)

	if types.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	types.LOG_DEBUG.Println("Browsing path ", accessPath.RealPath)
	counter := 0
	var result []types.StorageItem

	command.Browser.List.Result.CurrentItem = types.StorageItem{Name: accessPath.FileInfo.Name(), IsDir: accessPath.IsDir, ModificationDate: accessPath.FileInfo.ModTime().Unix(), Access: accessPath.Access, ShareAccess: accessPath.Access, Kind: filepath.Ext(accessPath.FileInfo.Name())}
	if (accessPath.IsDir) {
		fileList, err := ioutil.ReadDir(*accessPath.RealPath)
		if nil != err {
			types.LOG_ERROR.Println("Failed to read path with error code " + err.Error())
			resp <- types.EnumCommandHandlerError
		}
		result = make([]types.StorageItem, len(fileList))

		var access types.AccessType
		for _, file := range fileList {
			s := types.StorageItem{Name: file.Name(), IsDir: file.IsDir(), ModificationDate: file.ModTime().Unix()}
			if !file.IsDir() {
				s.Size = file.Size()
				s.Kind = filepath.Ext(file.Name())
				access = accessPath.Access
			} else {
				s.Kind = "folder"
				accessPath2, _ := auth.GetAccessAndPath(b.config, context, path.Join(command.Browser.List.Path, s.Name), asUser)
				if types.ERROR_NO_ERROR != accessPath2.Error {
					err = errors.New("Couldn't get infos about this")
				} else {
					err = nil
				}
				access = accessPath2.Access
			}
			if nil != err {
				continue
			} else if types.NONE == access && asUser {
				continue
			}
			s.Access = access
			s.ShareAccess = s.Access
			result[counter] = s
			counter++
		}
	}
	command.Browser.List.Result.Children = result[:counter]
	time.Sleep(2)
	resp <- types.EnumCommandHandlerDone
}

func (b *BrowseHandler) GetUploadPath(context *types.CommandContext) (path *string, size int64, hErr *types.HttpError) {
	command := context.Command
	if types.EnumBrowserUploadFile != command.Name {
		return nil, 0, &types.HttpError{errors.New("Not Allowed for this command type"), http.StatusBadRequest}
	}

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.UploadFile.Path, false)

	if types.ERROR_NO_ERROR != accessPath.Error {
		return nil, 0, &types.HttpError{errors.New("Access Error"), http.StatusUnauthorized}
	}
	if types.READ_WRITE != accessPath.Access {
		return nil, 0, &types.HttpError{errors.New("Access Error"), http.StatusUnauthorized}
	}

	return accessPath.RealPath, command.Browser.UploadFile.Size, nil
}
