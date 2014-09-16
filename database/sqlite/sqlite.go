package sqlite

import (
	"encoding/json"
	"errors"
	"fmt"
	// _ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	// _ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/types"
	"strconv"
	"strings"
	"sync"
)

const (
	Name string = "SqliteDb"
)

type UserSpecificAuth struct {
	AuthType string `sql:"type:varchar;"`
	Blob     string `sql:"type:varchar;"`
	UserId   string `sql:"type:varchar;"`
}

type User struct {
	Login   string `sql:"type:varchar;"`
	Email   string `sql:"type:varchar;"`
	Id      string `sql:"type:varchar;"` //This id should be unique depending on the DbType
	IsAdmin bool   `sql:"type:boolean;"` //Can only be changed by the admin
}

type pathAccesses struct {
	Path   string         `sql:"type:varchar;"`
	Access api.AccessType `sql:"type:integer;"`
	UserId string         `sql:"type:varchar;"`
}

type Share struct {
	User string  `sql:"type:varchar;"` //This will only be set by server. This is the user that issued the share link
	Name *string `sql:"type:varchar;"` // Name used for displaying the share link if multiple share links available
	Path *string `sql:"type:varchar;"` // Can be empty only if ShareLinkKey is provided
	Key  *string `sql:"type:varchar;"` // Can be empty only for a creation or on a Get
	// UserList *[]string             `sql:"type:varchar;"` // This is only available for EnumRestricted mode
	Type        api.EnumShareLinkType `sql:"type:integer;"`
	Access      api.AccessType        `sql:"type:integer;"` // What access would people coming with this link have
	Id          string                `sql:"type:varchar;"`
	Password    *string               `sql:"type:varchar;"`
	NbDownloads *int                  `sql:"type:integer;"` // Number of downloads for a file. This is only valid for file shared, not directories
}

type AllowedUserShare struct {
	User    string  `sql:"type:varchar;"`
	ShareId string  `sql:"type:varchar;"`
	Key     *string `sql:"type:varchar;"` // Can be empty only for a creation or on a Get
}

type SqliteDatabase struct {
	Database     string `json:"database"`
	db           gorm.DB
	lock         sync.RWMutex
	commandsList []*types.Command
	commandIndex int
}

func NewSqliteDatase(config *json.RawMessage) (d *SqliteDatabase, err error) {
	d = new(SqliteDatabase)
	if err = json.Unmarshal(*config, d); nil != err {
		return nil, err
	}
	db, err := gorm.Open("sqlite3", d.Database)
	d.commandsList = make([]*types.Command, 10)
	d.commandIndex = 0
	db.LogMode(true)
	db.DB().SetMaxIdleConns(10)

	// Create tables if not exist
	values := []interface{}{&types.Command{}, &types.DownloadLink{}, &UserSpecificAuth{}, &User{}, &types.Session{}, &Share{}, &AllowedUserShare{}, &pathAccesses{}}
	for _, value := range values {
		if db.HasTable(value) != true {
			if err := db.CreateTable(value).Error; err != nil {
				panic(fmt.Sprintf("No error should happen when create table, but got %+v", err))
			}
		}
	}
	d.db = db
	return d, nil
}

func (d *SqliteDatabase) Name() string {
	return Name
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

func (d *SqliteDatabase) Log(level LogLevel, message string) {
	switch level {
	case DEBUG:
		tools.LOG_DEBUG.Println("SqliteDb: ", message)
	case INFO:
		tools.LOG_INFO.Println("SqliteDb: ", message)
	case WARNING:
		tools.LOG_WARNING.Println("SqliteDb: ", message)
	case ERROR:
		tools.LOG_ERROR.Println("SqliteDb: ", message)
	}
}

func (d *SqliteDatabase) SaveCommand(command *types.Command) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	//TODO register cmd in database
	if 0 == len(command.ApiCommand.CommandId) {
		d.commandsList[d.commandIndex] = command
		command.ApiCommand.CommandId = strconv.Itoa(d.commandIndex)
		command.CommandId = strconv.Itoa(d.commandIndex)
		d.commandIndex += 1
		if len(d.commandsList) == d.commandIndex {
			new_list := make([]*types.Command, len(d.commandsList)*2)
			for i, comm := range d.commandsList {
				new_list[i] = comm
			}
			d.commandsList = new_list
		}
	}
	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Command", command.ApiCommand.Name))
	return nil
}

func (d *SqliteDatabase) ListCommands(user *string, offset int, limit int, search_parameters *api.CommandsSearchParameters) ([]*types.Command, int, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	//TODO register cmd in database
	tempResult := make([]*types.Command, d.commandIndex) // Maximum size this could have
	nbResult := 0
	for _, elem := range d.commandsList[0:d.commandIndex] {
		if elem.User == user {
			tempResult[nbResult] = elem
			nbResult += 1
		}
	}
	return tempResult[0:nbResult], nbResult, nil
}

func (d *SqliteDatabase) GetCommand(ref string) (command *types.Command, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	//TODO register cmd in database
	command_id, err := strconv.ParseInt(ref, 0, 0)
	if nil != err {
		return nil, err
	}
	command = d.commandsList[command_id]
	return command, nil
}

func (d *SqliteDatabase) DeleteCommand(ref *string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	//TODO register cmd in database
	return nil
}

func (d *SqliteDatabase) AddDownloadLink(link *types.DownloadLink) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	err = d.db.Table("download_links").Create(link).Error
	if err != nil {
		return err
	}
	d.Log(DEBUG, fmt.Sprintf("%s: %s", "Saving download link", link))
	return nil
}

func (d *SqliteDatabase) GetDownloadLink(ref string) (link *types.DownloadLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result types.DownloadLink
	err = d.db.Table("download_links").Where("link = ?", ref).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (d *SqliteDatabase) AddAccount(account *types.Account) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.addAccount(account)
}

func (d *SqliteDatabase) addAccount(account *types.Account) (err error) {
	//Check if same user already exists
	var count int
	err = d.db.Model(User{}).Table("users").Where("login = ?", account.Login).Or("email = ?", account.Email).Count(&count).Error
	if (err != nil) || (count != 0) {
		return errors.New("Account already exists")
	}
	var idx int
	err = d.db.Table("users").Count(&idx).Error
	if err != nil {
		return errors.New("Error during account creation")
	}

	user := User{Login: account.Login, Email: account.Email, IsAdmin: account.IsAdmin, Id: strconv.Itoa(idx + 1)}
	err = d.db.Table("users").Create(&user).Error
	if err != nil {
		return errors.New("Error during account creation")
	}

	for key, value := range account.Auths {
		authentification := UserSpecificAuth{AuthType: key, Blob: value.Blob, UserId: user.Id}
		err = d.db.Table("user_specific_auths").Create(&authentification).Error
		if err != nil {
			return errors.New("Error during account creation")
		}
	}

	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Account", account))

	return nil
}

func (d *SqliteDatabase) GetAccount(authType string, ref string) (account *types.Account, id string, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getAccount(authType, ref)
}

func (d *SqliteDatabase) getAccount(authType string, ref string) (account *types.Account, id string, err error) {
	var result User
	err = d.db.Table("users").Where("login = ?", ref).Or("email = ?", ref).First(&result).Error
	if err != nil {
		message := fmt.Sprintf("GetAccount: Couldn't find the desired account %s:%s", authType, ref)
		d.Log(ERROR, message)
		fmt.Println(err)
		return nil, "", errors.New(message)
	}
	user := types.Account{}
	user.Login = result.Login
	user.Email = result.Email
	user.Id = result.Id
	user.IsAdmin = result.IsAdmin
	user.Auths = make(map[string]types.AccountSpecificAuth)

	var auths []UserSpecificAuth
	err = d.db.Model(UserSpecificAuth{}).Table("user_specific_auths").Where("user_id = ?", result.Id).Find(&auths).Error
	if err != nil {
		return nil, "", errors.New("Error during account selection")
	}

	for _, auth := range auths {
		auth_specific := types.AccountSpecificAuth{}
		auth_specific.AuthType = auth.AuthType
		auth_specific.Blob = auth.Blob
		user.Auths[auth.AuthType] = auth_specific
	}
	if 0 == len(authType) {
		return &user, user.Id, nil
	}

	_, found := user.Auths[authType]
	if !found {
		message := fmt.Sprintf("Couldn't find this kind of authentication %s for %s", authType, ref)
		d.Log(ERROR, message)
		return nil, "", errors.New(message)
	}

	return &user, user.Id, nil
}

func (d *SqliteDatabase) UpdateAccount(id string, account *types.Account) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	account_id, err := strconv.ParseInt(id, 0, 0)
	if nil != err {
		return err
	}
	var user User
	err = d.db.Table("users").Where("id = ?", account_id).First(&user).Error
	if err != nil {
		message := fmt.Sprintf("Couldn't find user:%s", account_id)
		d.Log(ERROR, message)
		return errors.New(message)
	}

	user.Email = account.Email
	user.IsAdmin = account.IsAdmin
	// update login?
	err = d.db.Table("users").Save(&user).Error
	if err != nil {
		return errors.New("Update account error")
	}

	err = d.db.Table("user_specific_auths").Where("user_id = ?", account_id).Delete(UserSpecificAuth{}).Error
	if err != nil {
		return err
	}

	var auths []UserSpecificAuth
	err = d.db.Table("user_specific_auths").Where("user_id = ?", account_id).Find(&auths).Error
	if err != nil {
		return err
	}
	for key, value := range account.Auths {
		authentification := UserSpecificAuth{AuthType: key, Blob: value.Blob, UserId: user.Id}
		err = d.db.Table("user_specific_auths").Create(&authentification).Error
		if err != nil {
			return errors.New("Error during account creation")
		}
	}
	return nil
}

func (d *SqliteDatabase) GetUserAccount(id string) (account *types.Account, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result User
	err = d.db.Table("users").Where("id = ?", id).First(&result).Error
	if err != nil {
		message := fmt.Sprintf("GetAccount: Couldn't find the desired account %s", id)
		d.Log(ERROR, message)
		fmt.Println(err)
		return nil, errors.New(message)
	}
	user := types.Account{}
	user.Login = result.Login
	user.Email = result.Email
	user.Id = result.Id
	user.IsAdmin = result.IsAdmin
	user.Auths = make(map[string]types.AccountSpecificAuth)

	var auths []UserSpecificAuth
	err = d.db.Model(UserSpecificAuth{}).Table("user_specific_auths").Where("user_id = ?", result.Id).Find(&auths).Error
	if err != nil {
		return nil, errors.New("Error during account selection")
	}
	for _, auth := range auths {
		auth_specific := types.AccountSpecificAuth{}
		auth_specific.AuthType = auth.AuthType
		auth_specific.Blob = auth.Blob
		user.Auths[auth.AuthType] = auth_specific
	}

	return &user, nil
}

func (d *SqliteDatabase) ListAccounts(searchDict map[string]string) (accounts []*types.Account, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	var results []User
	err = d.db.Table("users").Find(&results).Error
	if err != nil {
		return nil, err
	}
	users := make([]*types.Account, len(results))
	index := 0

	for _, result := range results {

		account := types.Account{}
		account.Login = result.Login
		account.Email = result.Email
		account.Id = result.Id
		account.IsAdmin = result.IsAdmin
		account.Auths = make(map[string]types.AccountSpecificAuth)

		var auths []UserSpecificAuth
		err = d.db.Model(UserSpecificAuth{}).Table("user_specific_auths").Where("user_id = ?", result.Id).Find(&auths).Error
		if err != nil {
			return nil, err
		}
		for _, auth := range auths {
			auth_specific := types.AccountSpecificAuth{}
			auth_specific.AuthType = auth.AuthType
			auth_specific.Blob = auth.Blob
			account.Auths[auth.AuthType] = auth_specific
		}
		users[index] = &account
		index += 1

	}

	if 0 == len(searchDict) {
		d.Log(DEBUG, "No search parameters")
		return accounts, nil
	}

	d.Log(DEBUG, fmt.Sprintf("We had some search parameters ", searchDict))
	values := make([]*types.Account, len(results))
	i := 0
	for _, account := range users {
		for k, v := range searchDict {
			switch k {
			case "login":
				if strings.Contains(account.Login, v) {
					values[i] = account
					i += 1
					break
				}
			case "email":
				if strings.Contains(account.Email, v) {
					values[i] = account
					i += 1
					break
				}
			case "id":
				if strings.Contains(account.Id, v) {
					values[i] = account
					i += 1
					break
				}
			}
		}
	}
	return values[0:i], nil
}

func (d *SqliteDatabase) StoreSession(session *types.Session) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var count int
	err = d.db.Table("sessions").Where("user_id = ?", session.UserId).Count(&count).Error
	if err != nil {
		return errors.New("Error during account creation")
	}
	if count != 0 {
		return nil
	}
	err = d.db.Table("sessions").Create(session).Error
	if err != nil {
		return errors.New("Error during account creation")
	}
	return nil
}

func (d *SqliteDatabase) GetSession(ref string) (session *types.Session, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result types.Session
	err = d.db.Table("sessions").Where("authentication_header = ?", ref).First(&result).Error
	if err != nil {
		message := fmt.Sprintf("GetSession: Couldn't find the desired session %s", ref)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}
	return &result, nil
}

func (d *SqliteDatabase) RemoveSession(ref string) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	err = d.db.Table("sessions").Where("authentication_header = ?", ref).Delete(types.Session{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *SqliteDatabase) SaveShareLink(shareLink *types.ShareLink) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	//Todo check if there is already a sharelink with this name and user
	var idx int
	err = d.db.Table("shares").Count(&idx).Error
	if err != nil {
		return errors.New("Error during account creation")
	}
	access := *shareLink.ShareLink.Access

	share := Share{Name: shareLink.ShareLink.Name, Path: shareLink.ShareLink.Path,
		Key: shareLink.ShareLink.Key, User: shareLink.User,
		Id: strconv.Itoa(idx + 1), Type: shareLink.ShareLink.Type,
		Access: access, Password: shareLink.ShareLink.Password,
		NbDownloads: shareLink.ShareLink.NbDownloads}
	err = d.db.Table("shares").Create(&share).Error
	if err != nil {
		return errors.New("Register SaveShareLink error")
	}
	if shareLink.ShareLink.Type == api.EnumRestricted && shareLink.ShareLink.UserList != nil {
		for _, value := range *shareLink.ShareLink.UserList {
			user := AllowedUserShare{User: value, ShareId: strconv.Itoa(idx + 1), Key: shareLink.ShareLink.Key}
			err = d.db.Table("allowed_user_shares").Create(&user).Error
			if err != nil {
				return errors.New("Register SaveShareLink error")
			}
		}
	}

	return nil
}

func (d *SqliteDatabase) UpdateShareLink(shareLink *types.ShareLink) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	var share Share
	err = d.db.Table("shares").Where("path = ?", *shareLink.ShareLink.Path).First(&share).Error
	if err != nil {
		message := fmt.Sprintf("Couldn't find share link:%s", *shareLink.ShareLink.Path)
		d.Log(ERROR, message)
		return errors.New(message)
	}
	previous_type := share.Type
	share.Name = shareLink.ShareLink.Name
	share.Key = shareLink.ShareLink.Key
	share.Type = shareLink.ShareLink.Type
	share.Access = *shareLink.ShareLink.Access
	share.Password = shareLink.ShareLink.Password
	share.NbDownloads = shareLink.ShareLink.NbDownloads
	err = d.db.Table("shares").Save(&share).Error
	if err != nil {
		return errors.New("Update SaveShareLink error")
	}

	if previous_type == api.EnumRestricted && share.Type != api.EnumRestricted {
		// delete user share list
		err = d.db.Table("allowed_user_shares").Where("share_id = ?", share.Id).Delete(AllowedUserShare{}).Error
		if err != nil {
			return errors.New("Update SaveShareLink error")
		}
	}
	if previous_type != api.EnumRestricted && share.Type == api.EnumRestricted {
		//add user share list
		if shareLink.ShareLink.UserList != nil {
			for _, value := range *shareLink.ShareLink.UserList {
				user := AllowedUserShare{User: value, ShareId: share.Id, Key: shareLink.ShareLink.Key}
				err = d.db.Table("allowed_user_shares").Create(&user).Error
				if err != nil {
					return errors.New("Register SaveShareLink error")
				}
			}
		}
	}
	if previous_type == api.EnumRestricted && share.Type == api.EnumRestricted {
		//update user share list
		err = d.db.Table("allowed_user_shares").Where("share_id = ?", share.Id).Delete(AllowedUserShare{}).Error
		if err != nil {
			return errors.New("Update SaveShareLink error")
		}
		if shareLink.ShareLink.UserList != nil {
			for _, value := range *shareLink.ShareLink.UserList {
				user := AllowedUserShare{User: value, ShareId: share.Id, Key: shareLink.ShareLink.Key}
				err = d.db.Table("allowed_user_shares").Create(&user).Error
				if err != nil {
					return errors.New("Update SaveShareLink error")
				}
			}

		}
	}

	return nil
}

func (d *SqliteDatabase) GetShareLink(key string) (shareLink *types.ShareLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var share Share
	err = d.db.Table("shares").Where("key = ?", key).First(&share).Error
	if err != nil {
		message := fmt.Sprintf("Couldn't find share link %s", key)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}

	var users []AllowedUserShare
	var allowed_user []string
	if share.Type == api.EnumRestricted {
		err = d.db.Model(AllowedUserShare{}).Table("allowed_user_shares").Where("key = ?", key).Find(&users).Error
		if err != nil {
			message := fmt.Sprintf("Couldn't find share link %s", key)
			d.Log(ERROR, message)
			return nil, errors.New(message)
		}
		allowed_user = make([]string, len(users))
		idx := 0
		for _, user := range users {
			allowed_user[idx] = user.User
			idx += 1
		}
	}

	link := api.ShareLink{
		Name:        share.Name,
		Path:        share.Path,
		Key:         share.Key,
		UserList:    &allowed_user,
		Type:        share.Type,
		Access:      &share.Access,
		NbDownloads: share.NbDownloads,
		Password:    share.Password,
	}
	sharelink := types.ShareLink{
		User:      share.User,
		Id:        *share.Key,
		ShareLink: link,
	}
	return &sharelink, nil
}

func (d *SqliteDatabase) RemoveShareLink(key string) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	err = d.db.Table("shares").Where("key = ?", key).Delete(Share{}).Error
	if err != nil {
		return err
	}
	d.db.Table("allowed_user_shares").Where("key = ?", key).Delete(AllowedUserShare{})
	return nil
}

func (d *SqliteDatabase) ListShareLinks(user string) (shareLinks []*types.ShareLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.listShareLinks(user)

}

func (d *SqliteDatabase) listShareLinks(user string) (shareLinks []*types.ShareLink, err error) {

	var results []Share
	err = d.db.Table("shares").Where("user = ?", user).Find(&results).Error
	if err != nil {
		message := fmt.Sprintf("Couldn't find share link for user %s", user)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}
	shares := make([]*types.ShareLink, len(results))
	index := 0

	for _, result := range results {
		var users []AllowedUserShare
		var allowed_user []string
		if result.Type == api.EnumRestricted {
			d.db.Model(AllowedUserShare{}).Table("allowed_user_shares").Where("key = ?", result.Key).Find(&users)
			allowed_user = make([]string, len(users))
			idx := 0
			for _, user := range users {
				allowed_user[idx] = user.User
				idx += 1
			}
		}
		link := api.ShareLink{
			Name:        result.Name,
			Path:        result.Path,
			Key:         result.Key,
			UserList:    &allowed_user,
			Type:        result.Type,
			Access:      &result.Access,
			NbDownloads: result.NbDownloads,
			Password:    result.Password,
		}
		sharelink := types.ShareLink{
			User:      result.User,
			Id:        *result.Key,
			ShareLink: link,
		}
		shares[index] = &sharelink
		index += 1

	}
	return shares, nil
}

func (d *SqliteDatabase) GetShareLinksFromPath(path string, user string) (shareLink []*types.ShareLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getShareLinksFromPath(path, user)
}

func (d *SqliteDatabase) getShareLinksFromPath(path string, user string) (shareLinks []*types.ShareLink, err error) {
	shares, err := d.listShareLinks(user)
	if err != nil {
		message := fmt.Sprintf("Couldn't find share link for user %s path %s", user, path)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}

	shareLinks = make([]*types.ShareLink, 0, 10)
	var count = 0
	for _, shareLink := range shares {
		if *shareLink.ShareLink.Path == path {
			shareLinks = shareLinks[0 : count+1]
			shareLinks[count] = shareLink
			count += 1
			if count == cap(shareLinks) {
				newSlice := make([]*types.ShareLink, len(shareLinks), 2*cap(shareLinks))
				for i := range shareLinks {
					newSlice[i] = shareLinks[i]
				}
				shareLinks = newSlice
			}
		}
	}
	return shareLinks, nil
}

func (d *SqliteDatabase) GetAccess(user *string, path string) (api.AccessType, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getAccess(user, path)
}

func (d *SqliteDatabase) getAccess(user *string, path string) (api.AccessType, error) {
	var accesses []pathAccesses
	var user_id string
	if nil == user {
		user_id = ""
	} else {
		user_id = *user
	}
	err := d.db.Table("path_accesses").Where("user_id = ?", user_id).Find(&accesses).Error
	if err != nil {
		// If user is not Nil then look for the public accesses
		if nil == user {
			return api.NONE, nil
		}
		err = d.db.Table("path_accesses").Where("user_id = ?", "").Find(&accesses).Error
		if err != nil {
			return api.NONE, nil
		}
	}

	// Now check for the path
	splittedPath := strings.Split(path, "/")
	finalAccessType := api.NONE
	for i := len(splittedPath); i > 0; i-- {
		for _, access := range accesses {
			tools.LOG_DEBUG.Println("Looking for access to ", strings.Join(splittedPath[:i], "/"))
			tools.LOG_DEBUG.Println("Checking with ", access.Path, access.Access, access.UserId)
			if strings.Contains(strings.Join(splittedPath[:i], "/"), access.Path) {
				finalAccessType = access.Access
				tools.LOG_DEBUG.Println("access ", finalAccessType)
				break
			}
		}
		if finalAccessType != api.NONE {
			break
		}
	}
	if api.NONE == finalAccessType && nil != user {
		return d.getAccess(nil, path)
	}

	return finalAccessType, nil
}

func (d *SqliteDatabase) SetAccess(user *string, path string, access api.AccessType) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	var user_id string
	if nil == user {
		user_id = ""
	} else {
		user_id = *user
	}
	//check if access already exists
	var count int
	err := d.db.Model(User{}).Table("path_accesses").Where("user_id = ?", *user).Where("path = ?", path).Count(&count).Error
	if (err != nil) || (count != 0) {
		return errors.New("Access already exists")
	}
	path_access := pathAccesses{UserId: user_id, Path: path, Access: access}
	err = d.db.Table("path_accesses").Create(&path_access).Error
	if err != nil {
		message := fmt.Sprintf("Couldn't insert access %s path %s", user_id, path)
		d.Log(ERROR, message)
		return errors.New(message)
	}
	return nil
}

func (d *SqliteDatabase) ClearAccesses() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.db.DropTableIfExists(pathAccesses{})
	return nil
}
