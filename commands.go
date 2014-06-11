package main

import (
    "net/http"
	"encoding/json"
	"io/ioutil"
	"io"
	"fmt"
)

type CommandHandler struct{
	config *Configuration
}

func (c *CommandHandler)Init(config *Configuration){
	c.config = config
}

//Handle Request on /commands
//Only GET and POST request are available
func (c *CommandHandler)Commands(w http.ResponseWriter, r *http.Request){
	command := new(Command)
	input, err := ioutil.ReadAll(r.Body)
	if (nil != err) {
		fmt.Println("1 Failed with error code "+ err.Error())
		return
	}
	err = json.Unmarshal(input, command)
	//TODO start a timer or something like this so that we can timeout the request
	if (command.Name == EnumBrowse){
		fileList, err := ioutil.ReadDir(c.config.RootPrefix + command.Browse.Path)
		if (nil != err){
			fmt.Println("2 Failed with error code "+ err.Error())
			return
		}
		var result = make([]StorageItem, len(fileList))
		for i, file := range fileList {
			s := StorageItem{Name:file.Name(), IsDir:file.IsDir(), ModificationDate:file.ModTime()}
			if (!file.IsDir()){
				s.Size = file.Size()
			}
			result[i] = s
		}
		command.Browse.Results = result
	}
	command.State.Status = DONE
    b, _ := json.Marshal(command)
    io.WriteString(w, string(b))
}



func (c *CommandHandler)Command(w http.ResponseWriter, r *http.Request){

}

type EnumAction string

const (
	EnumBrowserBrowse EnumAction = "browser.browse"
	EnumBrowserCreateFolder EnumAction = "browser.create_folder"
	EnumBrowserDeleteItem EnumAction = "browser.delete_item"
	EnumDebugLongRequest EnumAction = "debug.long_request"
)

const (
	DONE = 0
	QUEUED = 1
	IN_PROGRESS = 2
	ERROR = 3
	CANCELLED = 4
)

type CommandStatus struct{
	Status int `json:"status"`
	Progress *int `json:"progress,omitempty"`
	ErrorCode *int `json:"error_code,omitempty"`
}

type StorageItem struct{
	Name string `json:"name"`
	IsDir bool `json:"isDir"`
	ModificationDate int `json:"mDate"`
	ChangeDate int `json:"cDate"`
	Size int64 `json:"size"`
}

type CommandBrowse struct {
	Path string `json:"path"`
	Results []StorageItem `json:"results"`
}

type CommandCreateFolder struct {
	Path string `json:"path"`
	Result StorageItem `json:"result"`
}

type CommandDeleteItem struct {
	Path string `json:"path"`
}

type Command struct {
	Name EnumAction `json:"name"` // Name of action Requested
	CommandId string `json:"command_id"` // Command Id returned by client when timeout is reached
	State CommandStatus `json:"state"`
	Timeout int `json:"timeout"` // Result should be returned before timeout, or client will have to poll using CommandId
	Browse *CommandBrowse `json:"browse_command,omitempty"`
	Delete *CommandDeleteItem `json:"delete_command,omitempty"`
	CreateFolder *CommandCreateFolder `json:"create_command,omitempty"`
}
