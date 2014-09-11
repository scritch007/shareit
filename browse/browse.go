package browse

import (
	"errors"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/auth"
	"github.com/scritch007/shareit/thumbnail"
	"github.com/scritch007/shareit/types"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

type BrowseHandler struct {
	config    *types.Configuration
	thumbnail *thumbnail.ThumbnailGenerator
}

func NewBrowseHandler(config *types.Configuration) (handler *BrowseHandler) {
	handler = &BrowseHandler{config: config, thumbnail: thumbnail.NewThumbnailGenerator(config)}
	return handler
}

func (b *BrowseHandler) Handle(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) *types.HttpError {
	command := context.Command
	tools.LOG_DEBUG.Println("Got this command ", command.ApiCommand.Name)
	if nil == command.ApiCommand.Browser {
		return &types.HttpError{Err: errors.New("Missing browse command body"), Status: http.StatusBadRequest}
	}
	tools.LOG_DEBUG.Println("Got this command ", command.ApiCommand.Name)
	if command.ApiCommand.Name == api.EnumBrowserList {
		go b.browseCommand(context, resp)
	} else if command.ApiCommand.Name == api.EnumBrowserCreateFolder {
		go b.createFolderCommand(context, resp)
	} else if command.ApiCommand.Name == api.EnumBrowserDelete {
		go b.deleteItemCommand(context, resp)
	} else if command.ApiCommand.Name == api.EnumBrowserDownloadLink {
		go b.downloadLink(context, resp)
	} else if command.ApiCommand.Name == api.EnumBrowserUploadFile {
		go b.uploadFile(context, resp)
	} else if command.ApiCommand.Name == api.EnumBrowserThumbnail {
		go b.thumbnailCommand(context, resp)
	} else {
		return &types.HttpError{Err: errors.New("Unknown Browse command"), Status: http.StatusBadRequest}
	}
	return nil
}

func (b *BrowseHandler) downloadLink(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command.ApiCommand
	if nil == command.Browser.DownloadLink {
		tools.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = api.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.DownloadLink.Input.Path, false)
	if !accessPath.Exists {
		tools.LOG_DEBUG.Println("Couldn't find this path")
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	file_path := accessPath.RealPath
	result := tools.ComputeHmac256(*file_path, b.config.PrivateKey)
	dLink := types.DownloadLink{Link: result, Path: command.Browser.DownloadLink.Input.Path, RealPath: file_path}
	b.config.Db.AddDownloadLink(&dLink)
	command.Browser.DownloadLink.Output.DownloadLink = url.QueryEscape(result)
	resp <- types.EnumCommandHandlerDone
}

//Handle removal of an item
func (b *BrowseHandler) deleteItemCommand(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command.ApiCommand
	if nil == command.Browser.Delete {
		tools.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = api.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	asUser := nil == context.Account || !context.Account.ApiAccount.IsAdmin

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.Delete.Input.Path, asUser)

	if api.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	if api.READ_WRITE != accessPath.Access {
		command.State.ErrorCode = api.ERROR_NOT_ALLOWED
		resp <- types.EnumCommandHandlerError
		return
	}

	item_path := accessPath.RealPath
	if nil == item_path {
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	if !accessPath.Exists {
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}
	if accessPath.IsDir {
		tools.LOG_DEBUG.Println("Item is a directory")
		//We are going to make something nice with a progress
		fileList, err := ioutil.ReadDir(*item_path)
		if nil != err {
			tools.LOG_DEBUG.Println("Couldn't list directory")
			command.State.ErrorCode = api.ERROR_FILE_SYSTEM
			resp <- types.EnumCommandHandlerError
			return
		}
		nbElements := len(fileList)
		success := types.EnumCommandHandlerDone
		for i, element := range fileList {
			element_path := path.Join(*item_path, element.Name())
			tools.LOG_DEBUG.Println("Trying to remove " + element_path)
			err = os.RemoveAll(element_path)
			if nil != err {
				success = types.EnumCommandHandlerError
				command.State.ErrorCode = api.ERROR_FILE_SYSTEM
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
	command := context.Command.ApiCommand
	if nil == command.Browser.CreateFolder {
		tools.LOG_ERROR.Println("Missing configuration")
		command.State.ErrorCode = api.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	asUser := nil == context.Account || !context.Account.ApiAccount.IsAdmin
	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.CreateFolder.Input.Path, asUser)

	if api.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	if api.READ_WRITE != accessPath.Access {
		command.State.ErrorCode = api.ERROR_NOT_ALLOWED
		resp <- types.EnumCommandHandlerError
		return
	}

	item_path := accessPath.RealPath
	if nil == item_path {
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	if accessPath.Exists {
		command.State.ErrorCode = api.ERROR_INVALID_PATH
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
	command := context.Command.ApiCommand
	if nil == command.Browser.UploadFile {
		tools.LOG_ERROR.Println("Missing configuration")
		command.State.ErrorCode = api.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}

	asUser := nil == context.Account || !context.Account.ApiAccount.IsAdmin

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.UploadFile.Input.Path, asUser)

	if api.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	item_path := accessPath.RealPath
	if nil == item_path {
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	//If write access has been removed, then we should respond false!!
	if api.READ_WRITE != accessPath.Access {
		command.State.ErrorCode = api.ERROR_NOT_ALLOWED
		resp <- types.EnumCommandHandlerError
		return
	}

	if accessPath.Exists {
		tools.LOG_DEBUG.Println("Truncating file")
		os.Truncate(*item_path, command.Browser.UploadFile.Input.Size)
		return
	}
	resp <- types.EnumCommandHandlerPostponed
}

func (b *BrowseHandler) thumbnailCommand(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command.ApiCommand
	if nil == command.Browser.Thumbnail {
		tools.LOG_ERROR.Println("Missing configuration")
		command.State.ErrorCode = api.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}

	asUser := nil == context.Account || !context.Account.ApiAccount.IsAdmin

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.Thumbnail.Input.Path, asUser)

	if api.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	item_path := accessPath.RealPath
	if nil == item_path {
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	//If write access has been removed, then we should respond false!!
	if api.NONE == accessPath.Access {
		command.State.ErrorCode = api.ERROR_NOT_ALLOWED
		resp <- types.EnumCommandHandlerError
		return
	}
	responseChan := make(chan string)
	thumbnailRequest := thumbnail.ThumbnailRequest{Path: *accessPath.RealPath, Response: responseChan}
	b.thumbnail.GetThumbnail(&thumbnailRequest)
	value := <-responseChan
	command.Browser.Thumbnail.Output.Content = value
	if 0 == len(value) {
		resp <- types.EnumCommandHandlerError
	} else {
		resp <- types.EnumCommandHandlerDone
	}
}

//Handle the browsing of a folder
func (b *BrowseHandler) browseCommand(context *types.CommandContext, resp chan<- types.EnumCommandHandlerStatus) {
	command := context.Command.ApiCommand
	if nil == command.Browser.List || 0 == len(command.Browser.List.Input.Path) {
		tools.LOG_ERROR.Println("Missing input configuration")
		command.State.ErrorCode = api.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}
	asUser := true
	if nil != context.Account && context.Account.ApiAccount.IsAdmin {
		//TODO checkt he asUser request header
		asUser = false
	}
	tools.LOG_DEBUG.Println("Retrieving access")
	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.List.Input.Path, asUser)

	if api.ERROR_NO_ERROR != accessPath.Error {
		command.State.ErrorCode = accessPath.Error
		resp <- types.EnumCommandHandlerError
		return
	}

	if !accessPath.Exists {
		command.State.ErrorCode = api.ERROR_INVALID_PATH
		resp <- types.EnumCommandHandlerError
		return
	}

	tools.LOG_DEBUG.Println("Browsing path ", accessPath.RealPath)
	counter := 0
	var result []api.StorageItem

	command.Browser.List.Output.CurrentItem = api.StorageItem{Name: filepath.Base(command.Browser.List.Input.Path), IsDir: accessPath.IsDir, MDate: accessPath.FileInfo.ModTime().Unix(), Access: accessPath.Access, ShareAccess: accessPath.Access, Kind: filepath.Ext(accessPath.FileInfo.Name()), Size: accessPath.FileInfo.Size(), Mimetype: mime.TypeByExtension(filepath.Ext(accessPath.FileInfo.Name()))}
	if accessPath.IsDir {
		fileList, err := ioutil.ReadDir(*accessPath.RealPath)
		if nil != err {
			tools.LOG_ERROR.Println("Failed to read path with error code " + err.Error())
			resp <- types.EnumCommandHandlerError
		}
		result = make([]api.StorageItem, len(fileList))

		var access api.AccessType
		for _, file := range fileList {
			s := api.StorageItem{Name: file.Name(), IsDir: file.IsDir(), MDate: file.ModTime().Unix()}
			if "." == string(file.Name()[0]) && (nil == command.Browser.List.Input.ShowHiddenFiles || !*command.Browser.List.Input.ShowHiddenFiles) {
				continue
			}
			if !file.IsDir() {
				s.Size = file.Size()
				s.Kind = filepath.Ext(file.Name())
				s.Mimetype = mime.TypeByExtension(filepath.Ext(file.Name()))
				access = accessPath.Access
			} else {
				s.Kind = "folder"
				s.Mimetype = ""
				accessPath2, _ := auth.GetAccessAndPath(b.config, context, path.Join(command.Browser.List.Input.Path, s.Name), asUser)
				if api.ERROR_NO_ERROR != accessPath2.Error {
					err = errors.New("Couldn't get infos about this")
				} else {
					err = nil
				}
				access = accessPath2.Access
			}
			if nil != err {
				continue
			} else if api.NONE == access && asUser {
				continue
			}
			s.Access = access
			s.ShareAccess = s.Access
			result[counter] = s
			counter++
		}
	} else {
		//Force the name for the display
		command.Browser.List.Output.CurrentItem.Name = accessPath.FileInfo.Name()
	}
	command.Browser.List.Output.Children = result[:counter]
	time.Sleep(2)
	resp <- types.EnumCommandHandlerDone
}

func (b *BrowseHandler) GetUploadPath(context *types.CommandContext) (path *string, size int64, hErr *types.HttpError) {
	command := context.Command.ApiCommand
	if api.EnumBrowserUploadFile != command.Name {
		return nil, 0, &types.HttpError{errors.New("Not Allowed for this command type"), http.StatusBadRequest}
	}

	accessPath, _ := auth.GetAccessAndPath(b.config, context, command.Browser.UploadFile.Input.Path, false)

	if api.ERROR_NO_ERROR != accessPath.Error {
		return nil, 0, &types.HttpError{errors.New("Access Error"), http.StatusUnauthorized}
	}
	if api.READ_WRITE != accessPath.Access {
		return nil, 0, &types.HttpError{errors.New("Access Error"), http.StatusUnauthorized}
	}

	return accessPath.RealPath, command.Browser.UploadFile.Input.Size, nil
}
