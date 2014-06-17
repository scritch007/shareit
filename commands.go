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
	"strconv"
	"time"
)

//ServeContent(w ResponseWriter, req *Request, name string, modtime time.Time, content io.ReadSeeker)

func (c *CommandHandler) save(command *Command) {
	c.commandsList[c.commandIndex] = command
	command.CommandId = strconv.Itoa(c.commandIndex)
	c.commandIndex += 1
	if len(c.commandsList) == c.commandIndex {
		new_list := make([]*Command, len(c.commandsList)*2)
		for i, comm := range c.commandsList {
			new_list[i] = comm
		}
		c.commandsList = new_list
	}
}

// CommandHandler constructor
func NewCommandHandler(config *Configuration) (c *CommandHandler) {
	c = new(CommandHandler)
	c.config = config
	c.commandsList = make([]*Command, 10)
	c.commandIndex = 0
	c.downloadLinks = make(map[string]string)
	return c
}

func (c *CommandHandler) downloadLink(command *Command, resp chan<- bool) {
	if nil == command.GenerateDownloadLink {
		LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	file_path := path.Join(c.config.RootPrefix, command.GenerateDownloadLink.Path)
	result := ComputeHmac256(file_path, c.config.PrivateKey)
	command.GenerateDownloadLink.Result = url.QueryEscape(result)
	c.downloadLinks[result] = file_path
	resp <- true

}

//Handle removal of an item
func (c *CommandHandler) deleteItemCommand(command *Command, resp chan<- bool) {
	if nil == command.Delete {
		LOG_DEBUG.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	item_path := path.Join(c.config.RootPrefix, command.Delete.Path)
	LOG_DEBUG.Println("delete " + item_path)
	fileInfo, err := os.Lstat(item_path)
	if nil != err {
		command.State.ErrorCode = 1 //TODO
		resp <- false
		return
	}
	if fileInfo.IsDir() {
		LOG_DEBUG.Println("Item is a directory")
		//We are going to make something nice with a progress
		fileList, err := ioutil.ReadDir(item_path)
		if nil != err {
			LOG_DEBUG.Println("Couldn't list directory")
			command.State.ErrorCode = 1 //TODO
			resp <- false
			return
		}
		nbElements := len(fileList)
		success := true
		for i, element := range fileList {
			element_path := path.Join(item_path, element.Name())
			LOG_DEBUG.Println("Trying to remove " + element_path)
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
func (c *CommandHandler) createFolderCommand(command *Command, resp chan<- bool) {
	if nil == command.CreateFolder {
		fmt.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	error := os.Mkdir(path.Join(c.config.RootPrefix, command.CreateFolder.Path), os.ModePerm)
	if nil != error {
		resp <- false
	} else {
		resp <- true
	}
}

//Handle the browsing of a folder
func (c *CommandHandler) browseCommand(command *Command, resp chan<- bool) {
	if nil == command.Browse {
		fmt.Println("Missing input configuration")
		command.State.ErrorCode = 1
		resp <- false
		return
	}
	fileList, err := ioutil.ReadDir(path.Join(c.config.RootPrefix, command.Browse.Path))
	if nil != err {
		fmt.Println("2 Failed with error code " + err.Error())
		resp <- false
	}
	var result = make([]StorageItem, len(fileList))
	for i, file := range fileList {
		s := StorageItem{Name: file.Name(), IsDir: file.IsDir(), ModificationDate: file.ModTime().Unix()}
		if !file.IsDir() {
			s.Size = file.Size()
		}
		result[i] = s
	}
	command.Browse.Results = result
	time.Sleep(2)
	resp <- true
}

//Handle Request on /commands
//Only GET and POST request are available
func (c *CommandHandler) Commands(w http.ResponseWriter, r *http.Request) {
	if "GET" == r.Method {
		// We want to list the commands that have been already answered
		b, _ := json.Marshal(c.commandsList[0:c.commandIndex])
		io.WriteString(w, string(b))
		return
	}
	// Extract the POST body
	command := new(Command)
	c.save(command)
	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		fmt.Println("1 Failed with error code " + err.Error())
		return
	}
	err = json.Unmarshal(input, command)
	channel := make(chan bool)
	command.State.Progress = 0
	command.State.ErrorCode = 0
	command.State.Status = COMMAND_STATUS_IN_PROGRESS
	//TODO start a timer or something like this so that we can timeout the request
	if command.Name == EnumBrowserBrowse {
		go c.browseCommand(command, channel)
	} else if command.Name == EnumBrowserCreateFolder {
		go c.createFolderCommand(command, channel)
	} else if command.Name == EnumBrowserDeleteItem {
		go c.deleteItemCommand(command, channel)

	} else if command.Name == EnumBrowserDownloadLink {
		go c.downloadLink(command, channel)
	} else {
		return
	}
	timeout := time.Duration(command.Timeout)
	if 0 == timeout {
		timeout = 10
	}
	timer := time.NewTimer(timeout * time.Second)
	select {
	case a := <-channel:
		fmt.Println("Got answer from command")
		timer.Stop()
		if a {
			command.State.Status = COMMAND_STATUS_DONE
		} else {
			command.State.Status = COMMAND_STATUS_ERROR
		}
		command.State.Progress = 100
	case <-timer.C:
		fmt.Println("Timer just elapsed")
	}
	b, _ := json.Marshal(command)
	io.WriteString(w, string(b))
}

func (c *CommandHandler) Command(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	command_id, _ := strconv.ParseInt(vars["command_id"], 0, 0)
	b, _ := json.Marshal(c.commandsList[command_id])
	io.WriteString(w, string(b))
}

func (c *CommandHandler) Download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]
	LOG_DEBUG.Println("Request for downloading following file we've got ", file, c.downloadLinks)

	path, found := c.downloadLinks[file]
	if found {
		http.ServeFile(w, r, path)
	} else {
		io.WriteString(w, "Download link is unavailable. Try renewing link")
	}
}

type EnumAction string

const (
	EnumBrowserBrowse       EnumAction = "browser.browse"
	EnumBrowserCreateFolder EnumAction = "browser.create_folder"
	EnumBrowserDeleteItem   EnumAction = "browser.delete_item"
	EnumBrowserDownloadLink EnumAction = "browser.download_link"
	EnumDebugLongRequest    EnumAction = "debug.long_request"
)
