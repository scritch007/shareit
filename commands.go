package shareit

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	//	"strconv"
	"archive/zip"
	"errors"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/browse"
	"github.com/scritch007/shareit/share_link"
	"github.com/scritch007/shareit/types"
	"os"
	"strconv"
	"strings"
	"time"
)

//ServeContent(w ResponseWriter, req *Request, name string, modtime time.Time, content io.ReadSeeker)
//CommandHandler is used to keep information about issued commands
type CommandHandler struct {
	config          *types.Configuration
	shareLink       *share_link.ShareLinkHandler
	browser         *browse.BrowseHandler
	UploadChunkSize int64 `json:"upload_chunk_size"`
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
	c.UploadChunkSize = int64(config.UploadChunkSize)
	return c
}

func (c *CommandHandler) getHandler(command *api.Command) types.CommandHandler {
	if strings.Contains(string(command.Name), "browser.") {
		return c.browser
	} else if strings.Contains(string(command.Name), share_link.COMMAND_PREFIX) {
		return c.shareLink
	} else {
		return nil
	}
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
		var userName *string = nil
		if nil != user {
			userName = &user.Id
		}
		commands, _, err := c.config.Db.ListCommands(userName, 0, -1, nil)
		if nil != err {
			errMessage := fmt.Sprintf("Invalid Input: %s", err)
			tools.LOG_ERROR.Println(errMessage)
			http.Error(w, errMessage, http.StatusInternalServerError)
			return
		}
		b, _ := json.Marshal(commands)
		io.WriteString(w, string(b))
		return
	}
	// Extract the POST body
	command := new(api.Command)

	input, err := ioutil.ReadAll(r.Body)
	if nil != err {
		errMessage := fmt.Sprintf("1 Failed with error code: %s", err)
		tools.LOG_ERROR.Println(errMessage)
		http.Error(w, errMessage, http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(input, command)

	backendCommand := new(types.Command)
	backendCommand.ApiCommand = command
	if nil != user {
		backendCommand.User = &user.Id //Store current user
	} else {
		backendCommand.User = nil
	}
	if nil != err {
		//TODO Set erro Code
	}
	channel := make(chan types.EnumCommandHandlerStatus)
	command.State.Progress = 0
	command.State.ErrorCode = 0
	command.State.Status = api.COMMAND_STATUS_IN_PROGRESS
	err = c.save(backendCommand)
	if nil != err {
		http.Error(w, "Couldn't save this command", http.StatusInternalServerError)
		return
	}

	handler := c.getHandler(command)
	if nil == handler {
		http.Error(w, "Unknown Request Type", http.StatusBadRequest)
		return
	}
	commandContext := types.CommandContext{backendCommand, user, r}
	hErr := handler.Handle(&commandContext, channel)
	if nil != hErr {
		http.Error(w, hErr.Err.Error(), hErr.Status)
		return
	}
	timeout := time.Duration(command.Timeout)
	if 0 == timeout {
		timeout = 5
	}
	//timer := time.NewTimer(1)
	timer := time.NewTimer(timeout * time.Second)

	select {
	case a := <-channel:
		tools.LOG_DEBUG.Println("Got answer from command")
		timer.Stop()
		if types.EnumCommandHandlerDone == a {
			command.State.Status = api.COMMAND_STATUS_DONE
			command.State.Progress = 100
		} else if types.EnumCommandHandlerError == a {
			command.State.Status = api.COMMAND_STATUS_ERROR
			command.State.Progress = 100
		}
		c.save(backendCommand)
	case <-timer.C:
		tools.LOG_DEBUG.Println("Timer just elapsed")
		go func() {
			//Wait for the command to end
			a := <-channel
			if types.EnumCommandHandlerDone == a {
				command.State.Status = api.COMMAND_STATUS_DONE
				command.State.Progress = 100
			} else if types.EnumCommandHandlerError == a {
				command.State.Status = api.COMMAND_STATUS_ERROR
				command.State.Progress = 100
			}
			c.save(backendCommand)
		}()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, _ := json.Marshal(command)
	io.WriteString(w, string(b))
}

//This is extracted from the net/http/fs.go file
type httpRange struct {
	start, length int64
}

func parseRange(ra string, size int64) (*httpRange, error) {
	ra = strings.TrimSpace(ra)
	if ra == "" {
		return nil, errors.New("invalid range 1")
	}
	if !strings.HasPrefix(ra, "bytes") {
		return nil, errors.New("invalid range 1.1")
	}
	ra = ra[6:]
	i := strings.Index(ra, "-")
	if i < 0 {
		return nil, errors.New("invalid range 2")
	}
	start, endAndSize := strings.TrimSpace(ra[:i]), strings.TrimSpace(ra[i+1:])

	i = strings.Index(endAndSize, "/")

	end, rSizeStr := strings.TrimSpace(endAndSize[:i]), strings.TrimSpace(endAndSize[i+1:])

	value, err := strconv.ParseInt(rSizeStr, 10, 64)
	if err != nil {
		return nil, errors.New("invalid range 2.5")
	}
	rSize := value

	if rSize != size {
		return nil, errors.New("Invalid range 3")
	}

	var r httpRange
	value, err = strconv.ParseInt(start, 10, 64)
	if err != nil || value > size || value < 0 {
		return nil, errors.New("invalid range 4")
	}
	r.start = value
	value, err = strconv.ParseInt(end, 10, 64)
	if err != nil || r.start > value {
		return nil, errors.New("invalid range 5")
	}
	if value >= size {
		value = size - 1
	}
	r.length = value - r.start + 1
	return &r, nil
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
	if nil != err {
		http.Error(w, fmt.Sprintf("Couldn't get this command ref %s", ref), http.StatusBadRequest)
		return
	}
	if nil != command.User && (nil == user || *command.User != user.Id) {
		http.Error(w, "You are trying to access some resources that do not belong to you", http.StatusUnauthorized)
		return
	}

	if "GET" == r.Method {
		b, _ := json.Marshal(command)
		io.WriteString(w, string(b))
	} else if "PUT" == r.Method {

		if 100 == command.ApiCommand.State.Progress {
			http.Error(w, "Command already completed", http.StatusUnauthorized)
			return
		}
		// make a buffer to keep chunks that are read
		buf := make([]byte, 0, c.UploadChunkSize)
		total_size := r.ContentLength
		var chunk_offset int64
		var buf_dim int64
		chunk_offset = int64(0)
		for {
			rest := total_size - chunk_offset
			if rest == 0 {
				// EOF
				break
			}
			if rest > c.UploadChunkSize {
				buf_dim = c.UploadChunkSize
			} else {
				buf_dim = rest
			}

			read_size, err := io.ReadFull(r.Body, buf[:buf_dim])
			if nil != err {
				errMessage := fmt.Sprintf("1 Failed with error code: %s", err)
				tools.LOG_ERROR.Println(errMessage)
				http.Error(w, errMessage, http.StatusBadRequest)
				return
			}
			tools.LOG_DEBUG.Println("Received ", read_size, "bytes")
			h := c.getHandler(command.ApiCommand)
			commandContext := types.CommandContext{command, user, r}
			uploadPath, size, hErr := h.GetUploadPath(&commandContext)

			if nil != hErr {
				errMessage := fmt.Sprintf("Failed to get upload path with error code: %s", hErr.Err)
				tools.LOG_ERROR.Println(errMessage)
				http.Error(w, errMessage, hErr.Status)
				return
			}
			rangeHeader := r.Header.Get("Content-Range")

			if _, err := os.Stat(*uploadPath); err != nil {
				if os.IsNotExist(err) {
					fo, err := os.Create(*uploadPath)
					if nil != err {
						errMessage := fmt.Sprintf("Couldn't create File with error %s", err)
						http.Error(w, errMessage, http.StatusInternalServerError)
						return
					}
					fo.Close()
				} else {
					errMessage := fmt.Sprintf("Couldn't read stat with error %s", err)
					tools.LOG_ERROR.Println(errMessage)
					http.Error(w, errMessage, http.StatusInternalServerError)
					return
				}
			}
			var offset int64 = 0
			if 0 != len(rangeHeader) {
				rangeValue, err := parseRange(rangeHeader, size)
				if nil != err {
					errMessage := fmt.Sprintf("Incorrect Range header %s", err.Error())
					tools.LOG_ERROR.Println(errMessage)
					http.Error(w, errMessage, http.StatusBadRequest)
					return
				}
				if size < rangeValue.start {
					errMessage := fmt.Sprintf("Couldn't seek to requested offset %d", rangeValue.start)
					tools.LOG_ERROR.Println(errMessage)
					http.Error(w, errMessage, http.StatusBadRequest)
					return
				}

				offset = rangeValue.start
			}
			f, err := os.OpenFile(*uploadPath, os.O_RDWR, os.ModePerm)
			if nil != err {
				errMessage := fmt.Sprintf("Failed to open file with error %s", err)
				tools.LOG_ERROR.Println(errMessage)
				http.Error(w, errMessage, http.StatusInternalServerError)
				return
			}
			defer f.Close()
			f.Seek(offset+chunk_offset, os.SEEK_SET)
			io.WriteString(f, string(buf[:read_size]))
			command.ApiCommand.State.Progress = int((offset + chunk_offset + int64(read_size)) * 100 / size)
			if 100 == command.ApiCommand.State.Progress {
				command.ApiCommand.State.Status = api.COMMAND_STATUS_DONE
			}
			chunk_offset += int64(read_size)
		}
	}
}

func (c *CommandHandler) Download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	file := vars["file"]

	link, err := c.config.Db.GetDownloadLink(file)
	//Get the realpath depending on the configuration and the sharelink or direct download
	if nil == err {
		tools.LOG_DEBUG.Println("Serving file ", *link.RealPath)
		fileInfo, err := os.Lstat(*link.RealPath)
		if nil != err {
			http.Error(w, "Download link doesn't point to a valid path", http.StatusNotFound)
			return
		}
		if fileInfo.IsDir() {
			zipFileName := fileInfo.Name() + ".zip"
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", `attachment; filename="`+zipFileName+`"`)
			zw := zip.NewWriter(w)
			defer zw.Close()
			// Walk directory.
			filepath.Walk(*link.RealPath, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				// Remove base path, convert to forward slash.
				zipPath := path[len(*link.RealPath):]
				zipPath = strings.TrimLeft(strings.Replace(zipPath, `\`, "/", -1), `/`)
				ze, err := zw.Create(zipPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Cannot create zip entry <%s>: %s\n", zipPath, err)
					return err
				}
				file, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Cannot open file <%s>: %s\n", path, err)
					return err
				}
				defer file.Close()
				io.Copy(ze, file)
				return nil
			})

		} else {
			w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(*link.RealPath))
			http.ServeFile(w, r, *link.RealPath)
		}

	} else {
		http.Error(w, "Download link is unavailable. Try renewing link", http.StatusNotFound)
	}
}
