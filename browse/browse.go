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

func (b *BrowseHandler) Handle(command *types.Command, resp chan<- types.EnumCommandHandlerStatus) (error, int) {
	if nil == command.Browser {
		return errors.New("Missing browse command body"), http.StatusBadRequest
	}
	if command.Name == types.EnumBrowserBrowse {
		go b.browseCommand(command, resp)
	} else if command.Name == types.EnumBrowserCreateFolder {
		go b.createFolderCommand(command, resp)
	} else if command.Name == types.EnumBrowserDeleteItem {
		go b.deleteItemCommand(command, resp)
	} else if command.Name == types.EnumBrowserDownloadLink {
		go b.downloadLink(command, resp)
	} else if command.Name == types.EnumBrowserUploadFile {
		go b.uploadFile(command, resp)
	} else {
		return errors.New("Unknown Browse command"), http.StatusBadRequest
	}
	return nil, 0
}

func (b *BrowseHandler) downloadLink(command *types.Command, resp chan<- types.EnumCommandHandlerStatus) {
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
func (b *BrowseHandler) deleteItemCommand(command *types.Command, resp chan<- types.EnumCommandHandlerStatus) {
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
func (b *BrowseHandler) createFolderCommand(command *types.Command, resp chan<- types.EnumCommandHandlerStatus) {
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

func (b *BrowseHandler) uploadFile(command *types.Command, resp chan<- types.EnumCommandHandlerStatus) {
	if nil == command.Browser.UploadFile {
		types.LOG_ERROR.Println("Missing configuration")
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

	if nil != fileInfo {
		//This file already exists... (TODO CHECKTHAT)
		command.State.ErrorCode = types.ERROR_INVALID_PARAMETERS
		resp <- types.EnumCommandHandlerError
		return
	}
	resp <- types.EnumCommandHandlerPostponed
}

//Handle the browsing of a folder
func (b *BrowseHandler) browseCommand(command *types.Command, resp chan<- types.EnumCommandHandlerStatus) {
	if nil == command.Browser.List {
		types.LOG_ERROR.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- types.EnumCommandHandlerError
		return
	}

	//First check if we have a Key. If we do then we'll chroot the browse command...
	chroot := ""
	if nil != command.AuthKey {
		share_link, err := b.config.Db.GetShareLink(*command.AuthKey)
		if nil != err {
			command.State.ErrorCode = types.ERROR_INVALID_PARAMETERS
			resp <- types.EnumCommandHandlerError
			return
		}
		chroot = *share_link.Path
	}
	realPath := path.Join(b.config.RootPrefix, chroot, command.Browser.List.Path)
	types.LOG_DEBUG.Println("Browsing path ", realPath)
	fileList, err := ioutil.ReadDir(realPath)
	if nil != err {
		types.LOG_ERROR.Println("Failed to read path with error code " + err.Error())
		resp <- types.EnumCommandHandlerError
	}
	var result = make([]types.StorageItem, len(fileList))
	for i, file := range fileList {
		s := types.StorageItem{Name: file.Name(), IsDir: file.IsDir(), ModificationDate: file.ModTime().Unix()}
		if !file.IsDir() {
			s.Size = file.Size()
			s.Kind = filepath.Ext(file.Name())
		} else {
			s.Kind = "folder"
		}
		result[i] = s
	}
	command.Browser.List.Results = result
	time.Sleep(2)
	resp <- types.EnumCommandHandlerError
}

func (b *BrowseHandler) GetUploadPath(command *types.Command) (*string, error, int) {
	if types.EnumBrowserUploadFile != command.Name {
		return nil, errors.New("Not Allowed for this command type"), http.StatusBadRequest
	}
	item_path, _ := b.checkItemPath(&command.Browser.UploadFile.Path)
	if nil == item_path {
		return nil, errors.New("Invalid parameter"), http.StatusBadRequest
	}
	return item_path, nil, 0
}
