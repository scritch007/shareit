package shareit

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	//	"strconv"
	"github.com/scritch007/shareit/share_link"
	"github.com/scritch007/shareit/types"
	"strings"
	"time"
)

//ServeContent(w ResponseWriter, req *Request, name string, modtime time.Time, content io.ReadSeeker)
//CommandHandler is used to keep information about issued commands
type CommandHandler struct {
	config    *types.Configuration
	shareLink *share_link.ShareLinkHandler
}

func (c *CommandHandler) save(command *types.Command) error {
	return c.config.Db.SaveCommand(command)
}

// CommandHandler constructor
func NewCommandHandler(config *types.Configuration) (c *CommandHandler) {
	c = new(CommandHandler)
	c.config = config
	c.shareLink = share_link.NewShareLinkHandler(config)
	return c
}

func (c *CommandHandler) downloadLink(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.GenerateDownloadLink {
		types.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	file_path := path.Join(c.config.RootPrefix, command.Browser.GenerateDownloadLink.Path)
	result := ComputeHmac256(file_path, c.config.PrivateKey)
	dLink := types.DownloadLink{Link: result, Path: command.Browser.GenerateDownloadLink.Path}
	c.config.Db.AddDownloadLink(&dLink)
	command.Browser.GenerateDownloadLink.Result.Link = url.QueryEscape(result)
	command.Browser.GenerateDownloadLink.Result.Path = command.Browser.GenerateDownloadLink.Path
	resp <- true
}

//Handle removal of an item
func (c *CommandHandler) deleteItemCommand(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.Delete {
		types.LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	item_path := path.Join(c.config.RootPrefix, command.Browser.Delete.Path)
	types.LOG_DEBUG.Println("delete " + item_path)
	fileInfo, err := os.Lstat(item_path)
	if nil != err {
		command.State.ErrorCode = 1 //TODO
		resp <- false
		return
	}
	if fileInfo.IsDir() {
		types.LOG_DEBUG.Println("Item is a directory")
		//We are going to make something nice with a progress
		fileList, err := ioutil.ReadDir(item_path)
		if nil != err {
			types.LOG_DEBUG.Println("Couldn't list directory")
			command.State.ErrorCode = 1 //TODO
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
				command.State.ErrorCode = 1 //TODO
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
func (c *CommandHandler) createFolderCommand(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.CreateFolder {
		fmt.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	error := os.Mkdir(path.Join(c.config.RootPrefix, command.Browser.CreateFolder.Path), os.ModePerm)
	if nil != error {
		resp <- false
	} else {
		resp <- true
	}
}

//Handle the browsing of a folder
func (c *CommandHandler) browseCommand(command *types.Command, resp chan<- bool) {
	if nil == command.Browser.List {
		fmt.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	fileList, err := ioutil.ReadDir(path.Join(c.config.RootPrefix, command.Browser.List.Path))
	if nil != err {
		fmt.Println("2 Failed with error code " + err.Error())
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

//Handle Request on /commands
//Only GET and POST request are available
func (c *CommandHandler) Commands(w http.ResponseWriter, r *http.Request) {
	user, err := c.config.Auth.GetAuthenticatedUser(w, r)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if "GET" == r.Method {
		// We want to list the commands that have been already answered
		commands, _, err := c.config.Db.ListCommands(user, 0, -1, nil)
		if nil != err {
			errMessage := fmt.Sprintf("Invalid Input: %s", err)
			types.LOG_ERROR.Println(errMessage)
			http.Error(w, errMessage, http.StatusInternalServerError)
			return
		}
		b, _ := json.Marshal(commands)
		io.WriteString(w, string(b))
		return
	}
	// Extract the POST body
	command := new(types.Command)

	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		errMessage := fmt.Sprintf("1 Failed with error code: %s", err)
		types.LOG_ERROR.Println(errMessage)
		http.Error(w, errMessage, http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(input, command)
	command.User = user //Store current user
	if nil != err {
		//TODO Set erro Code
	}
	channel := make(chan bool)
	command.State.Progress = 0
	command.State.ErrorCode = 0
	command.State.Status = types.COMMAND_STATUS_IN_PROGRESS
	err = c.save(command)
	if nil != err {
		//TODO do something in that case
		return
	}
	if strings.Contains(string(command.Name), "browser.") {
		fmt.Println("In Browser...")
		if nil == command.Browser {
			http.Error(w, "Missing browse command body", http.StatusBadRequest)
			return
		}
		if command.Name == types.EnumBrowserBrowse {
			go c.browseCommand(command, channel)
		} else if command.Name == types.EnumBrowserCreateFolder {
			go c.createFolderCommand(command, channel)
		} else if command.Name == types.EnumBrowserDeleteItem {
			go c.deleteItemCommand(command, channel)
		} else if command.Name == types.EnumBrowserDownloadLink {
			go c.downloadLink(command, channel)
		} else {
			http.Error(w, "Missing browse command body", http.StatusBadRequest)
			return
		}
	} else if strings.Contains(string(command.Name), share_link.COMMAND_PREFIX) {
		go c.shareLink.Handle(command, channel)
	} else {
		http.Error(w, "Unknown Request Type", http.StatusBadRequest)
		return
	}
	timeout := time.Duration(command.Timeout)
	if 0 == timeout {
		timeout = 10
	}
	//timer := time.NewTimer(1)
	timer := time.NewTimer(timeout * time.Second)

	select {
	case a := <-channel:
		fmt.Println("Got answer from command")
		timer.Stop()
		if a {
			command.State.Status = types.COMMAND_STATUS_DONE
		} else {
			command.State.Status = types.COMMAND_STATUS_ERROR
		}
		command.State.Progress = 100
		c.save(command)
	case <-timer.C:
		fmt.Println("Timer just elapsed")
		go func() {
			//Wait for the command to end
			a := <-channel
			if a {
				command.State.Status = types.COMMAND_STATUS_DONE
			} else {
				command.State.Status = types.COMMAND_STATUS_ERROR
			}
			command.State.Progress = 100
			c.save(command)
		}()
	}
	b, _ := json.Marshal(command)
	io.WriteString(w, string(b))
}

func (c *CommandHandler) Command(w http.ResponseWriter, r *http.Request) {
	user, err := c.config.Auth.GetAuthenticatedUser(w, r)
	if nil != err {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	ref := vars["command_id"]
	command, err := c.config.Db.GetCommand(ref)
	if command.User != user {
		http.Error(w, "You are trying to access some resources that do not belong to you", http.StatusUnauthorized)
	}
	if nil != err {
		return
	}
	b, _ := json.Marshal(command)
	io.WriteString(w, string(b))
}

func (c *CommandHandler) Download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]

	link, err := c.config.Db.GetDownloadLink(file)
	if nil == err {
		http.ServeFile(w, r, path.Join(c.config.RootPrefix, link.Path))
	} else {
		http.Error(w, "Download link is unavailable. Try renewing link", http.StatusNotFound)
	}
}
