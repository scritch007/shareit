package browse

import (
	"github.com/scritch007/shareit/tools"
	"github.com/scritch007/shareit/types"
	"io/ioutil"
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

func (b *BrowseHandler) DownloadLink(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.GenerateDownloadLink {
		types.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- false
		return
	}
	file_path := path.Join(b.config.RootPrefix, command.Browser.GenerateDownloadLink.Path)
	result := tools.ComputeHmac256(file_path, b.config.PrivateKey)
	dLink := types.DownloadLink{Link: result, Path: command.Browser.GenerateDownloadLink.Path}
	b.config.Db.AddDownloadLink(&dLink)
	command.Browser.GenerateDownloadLink.Result.Link = url.QueryEscape(result)
	command.Browser.GenerateDownloadLink.Result.Path = command.Browser.GenerateDownloadLink.Path
	resp <- true
}

//Handle removal of an item
func (b *BrowseHandler) DeleteItemCommand(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.Delete {
		types.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- false
		return
	}
	item_path := path.Join(b.config.RootPrefix, command.Browser.Delete.Path)
	types.LOG_DEBUG.Println("delete " + item_path)
	fileInfo, err := os.Lstat(item_path)
	if nil != err {
		command.State.ErrorCode = types.ERROR_INVALID_PATH
		resp <- false
		return
	}
	if fileInfo.IsDir() {
		types.LOG_DEBUG.Println("Item is a directory")
		//We are going to make something nice with a progress
		fileList, err := ioutil.ReadDir(item_path)
		if nil != err {
			types.LOG_DEBUG.Println("Couldn't list directory")
			command.State.ErrorCode = types.ERROR_FILE_SYSTEM
			resp <- false
			return
		}
		nbElements := len(fileList)
		success := true
		for i, element := range fileList {
			element_path := path.Join(item_path, element.Name())
			types.LOG_DEBUG.Println("Trying to remove " + element_path)
			err = os.RemoveAll(element_path)
			if nil != err {
				success = false
				command.State.ErrorCode = types.ERROR_FILE_SYSTEM
			}
			command.State.Progress = i * 100 / nbElements
		}
		if nil != os.RemoveAll(item_path) {
			success = false
		}
		resp <- success
	} else {
		err = os.Remove(item_path)
		if nil == err {
			resp <- true
		} else {
			resp <- false
		}
	}

}

//Handle the creation of a folder
func (b *BrowseHandler) CreateFolderCommand(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.CreateFolder {
		types.LOG_ERROR.Println("Missing configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	error := os.Mkdir(path.Join(b.config.RootPrefix, command.Browser.CreateFolder.Path), os.ModePerm)
	if nil != error {
		resp <- false
	} else {
		resp <- true
	}
}

//Handle the browsing of a folder
func (b *BrowseHandler) BrowseCommand(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.List {
		types.LOG_ERROR.Println("Missing input configuration")
		command.State.ErrorCode = types.ERROR_MISSING_COMMAND_BODY
		resp <- false
		return
	}

	//First check if we have a Key. If we do then we'll chroot the browse command...
	chroot := ""
	if nil != command.AuthKey {
		share_link, err := b.config.Db.GetShareLink(*command.AuthKey)
		if nil != err {
			command.State.ErrorCode = types.ERROR_INVALID_PARAMETERS
			resp <- false
			return
		}
		chroot = *share_link.Path
	}
	realPath := path.Join(b.config.RootPrefix, chroot, command.Browser.List.Path)
	types.LOG_DEBUG.Println("Browsing path ", realPath)
	fileList, err := ioutil.ReadDir(realPath)
	if nil != err {
		types.LOG_ERROR.Println("Failed to read path with error code " + err.Error())
		resp <- false
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
	resp <- true
}
