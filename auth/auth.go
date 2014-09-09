package auth

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/auth/dummy"
	"github.com/scritch007/shareit/types"
	"os"
	"path"
	"path/filepath"
)

type configSubStruct struct {
	Type   string           `json:"type"`
	Config *json.RawMessage `json:"config"`
}

//Should be called by authentication mechanism
func NewAuthentication(config *json.RawMessage, r *mux.Router, globalConfig *types.Configuration) (newAuth *types.Authentication, err error) {
	var authConfigs []configSubStruct
	err = json.Unmarshal(*config, &authConfigs)
	newAuth = new(types.Authentication)
	newAuth.Config = globalConfig
	newAuth.Auths = make([]types.SubAuthentication, len(authConfigs))
	var newSubAuth types.SubAuthentication
	for i, elem := range authConfigs {
		switch elem.Type {
		case dummy.Name:
			newSubAuth, err = dummy.NewDummyAuth(elem.Config, globalConfig)
		default:
			err = errors.New("Unknown authentication method " + elem.Type)
			newAuth = nil
		}
		if nil != err {
			return nil, err
		}
		newSubAuth.AddRoutes(r)
		newAuth.Auths[i] = newSubAuth
	}

	return newAuth, err
}

type AccessPath struct {
	RealPath    *string
	Access      api.AccessType
	Error       api.EnumCommandErrorCode
	IsDir       bool
	Exists      bool
	IsShareLink bool
	FileInfo    os.FileInfo
}

func GetAccessAndPath(config *types.Configuration, context *types.CommandContext, inPath string, asUser bool) (AccessPath, error) {
	//First check if we have a Key. If we do then we'll chroot the browse command...
	chroot := ""
	access := api.READ // Default access type
	isRoot := "/" == inPath
	var accessPath = AccessPath{Access: api.NONE, Error: api.ERROR_NO_ERROR, Exists: false, IsShareLink: false, FileInfo: nil}

	if nil != context.Command.ApiCommand.AuthKey {
		tools.LOG_DEBUG.Println("There's an auth key")
		share_link, err := config.Db.GetShareLink(*context.Command.ApiCommand.AuthKey)
		accessPath.IsShareLink = true
		if nil != err {
			tools.LOG_ERROR.Println("Share link error " + err.Error())
			accessPath.Error = api.ERROR_INVALID_PARAMETERS
			return accessPath, err
		}
		chroot = *share_link.ShareLink.Path
		if nil != share_link.ShareLink.Access {
			access = *share_link.ShareLink.Access
		}
		//TODO add some check depending on the type of share_link...

		//Check if share_link is a directory if not check that basename/dirname are correct
		fileInfo, err := os.Lstat(path.Join(config.RootPrefix, chroot))
		if nil != err {
			tools.LOG_ERROR.Println(err)
			accessPath.Error = api.ERROR_INVALID_PATH
			return accessPath, err
		}
		if !fileInfo.IsDir() {
			//Force Access to readOnly
			accessPath.Access = api.READ
			baseName := filepath.Base(chroot)
			if "/" != inPath && baseName != inPath[1:] {
				tools.LOG_ERROR.Println(err)
				accessPath.Error = api.ERROR_INVALID_PATH
				return accessPath, err
			} else if "/" != inPath {
				chroot = filepath.Dir(chroot)
			}
		}
	} else {
		tools.LOG_DEBUG.Println("There's no auth key")
		if !asUser {
			access = api.READ_WRITE
		} else {
			if !isRoot {
				//Check if user has access to this path
				var err error
				access, err = config.Db.GetAccess(context.Command.User, inPath)
				if nil != err {
					tools.LOG_ERROR.Println("Couldn't get access " + err.Error())
					accessPath.Error = api.ERROR_INVALID_PATH
					return accessPath, err
				}
				if api.NONE == access {
					accessPath.Error = api.ERROR_NOT_ALLOWED
					return accessPath, nil
				}
			} else {
				if config.AllowRootWrite {
					//TODO
				}
			}
		}
	}
	realPath := path.Clean(path.Join(config.RootPrefix, chroot, inPath))
	tools.LOG_DEBUG.Println("Realpath is " + realPath)
	fileInfo, err := os.Lstat(realPath)
	if nil != err {
		if !os.IsNotExist(err) {
			tools.LOG_ERROR.Println("Error accessing to the file " + realPath + err.Error())
			accessPath.Error = api.ERROR_INVALID_PATH
			return accessPath, err
		}
	} else {
		accessPath.IsDir = fileInfo.IsDir()
		accessPath.Exists = true
		accessPath.FileInfo = fileInfo
	}

	accessPath.RealPath = &realPath
	accessPath.Access = access

	tools.LOG_DEBUG.Println("Realpath is ", realPath)
	return accessPath, nil
}
