package shareit

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	//	"strconv"
	"github.com/scritch007/shareit/browse"
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
	browser   *browse.BrowseHandler
}

func (c *CommandHandler) save(command *types.Command) error {
	return c.config.Db.SaveCommand(command)
}

// CommandHandler constructor
func NewCommandHandler(config *types.Configuration) (c *CommandHandler) {
	c = new(CommandHandler)
	c.config = config
	c.shareLink = share_link.NewShareLinkHandler(config)
	c.browser = browse.NewBrowseHandler(config)
	return c
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
		http.Error(w, "Couldn't save this command", http.StatusInternalServerError)
		return
	}
	if strings.Contains(string(command.Name), "browser.") {
		types.LOG_DEBUG.Println("Browser command")
		if nil == command.Browser {
			http.Error(w, "Missing browse command body", http.StatusBadRequest)
			return
		}
		if command.Name == types.EnumBrowserBrowse {
			go c.browser.BrowseCommand(command, channel)
		} else if command.Name == types.EnumBrowserCreateFolder {
			go c.browser.CreateFolderCommand(command, channel)
		} else if command.Name == types.EnumBrowserDeleteItem {
			go c.browser.DeleteItemCommand(command, channel)
		} else if command.Name == types.EnumBrowserDownloadLink {
			go c.browser.DownloadLink(command, channel)
		} else {
			http.Error(w, "Missing browse command body", http.StatusBadRequest)
			return
		}
	} else if strings.Contains(string(command.Name), share_link.COMMAND_PREFIX) {
		types.LOG_DEBUG.Println("Share Link command")
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
		types.LOG_DEBUG.Println("Got answer from command")
		timer.Stop()
		if a {
			command.State.Status = types.COMMAND_STATUS_DONE
		} else {
			command.State.Status = types.COMMAND_STATUS_ERROR
		}
		command.State.Progress = 100
		c.save(command)
	case <-timer.C:
		types.LOG_DEBUG.Println("Timer just elapsed")
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
