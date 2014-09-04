package mongo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/scritch007/ShareMinatorApiGenerator/api"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/types"
	// "io/ioutil"
	// "os"
	// "path"
	//"path/filepath"
	"strconv"
	"strings"
	"sync"
	//"github.com/scritch007/shareit/database"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	Name string = "MongoDb"
)

// type pathAccesses struct {
// Accesses map[string]api.AccessType
// }

type pathAccesses struct {
	Accesses map[string]api.AccessType `json:"access" bson:"access"`
	UserId   string                    `json:"user_id" bson:"user_id"`
}

type MongoDatabase struct {
	session *mgo.Session
	lock    sync.RWMutex
}

func NewMongoDatase(config *json.RawMessage) (d *MongoDatabase, err error) {
	d = new(MongoDatabase)
	// TODO set ip in config.json
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		fmt.Println("Unable to connect to local mongo instance!")
		return nil, err
	}
	d.session = session
	d.session.SetMode(mgo.Monotonic, true)
	return d, nil
}

func (d *MongoDatabase) Name() string {
	return Name
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

func (d *MongoDatabase) Log(level LogLevel, message string) {
	switch level {
	case DEBUG:
		tools.LOG_DEBUG.Println("MongoDb: ", message)
	case INFO:
		tools.LOG_INFO.Println("MongoDb: ", message)
	case WARNING:
		tools.LOG_WARNING.Println("MongoDb: ", message)
	case ERROR:
		tools.LOG_ERROR.Println("MongoDb: ", message)
	}
}

func (d *MongoDatabase) SaveCommand(command *types.Command) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("commands")

	if 0 == len(command.ApiCommand.CommandId) {
		index, err := c.Find(nil).Count()
		if err != nil {
			return errors.New("Register error")
		}
		command.ApiCommand.CommandId = strconv.Itoa(index + 1)
		command.CommandId = command.ApiCommand.CommandId
		err = c.Insert(command)
		if err != nil {
			return errors.New("Register error")
		}
	}

	err = c.Update(bson.M{"command_id": command.ApiCommand.CommandId}, command)
	if nil != err {
		return errors.New("Update error")
	}
	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Command", command.ApiCommand.Name))
	return nil
}

func (d *MongoDatabase) ListCommands(user *string, offset int, limit int, search_parameters *api.CommandsSearchParameters) ([]*types.Command, int, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var results []*types.Command
	c := d.session.DB("mongo").C("commands")
	// TODO: use search parameters?
	err := c.Find(bson.M{"user": user}).All(&results)
	if err != nil {
		return nil, 0, errors.New("I/O error in database")
	}

	return results, len(results), nil
}

func (d *MongoDatabase) GetCommand(ref string) (command *types.Command, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	c := d.session.DB("mongo").C("commands")

	command_id, err := strconv.ParseInt(ref, 0, 0)
	if nil != err {
		return nil, err
	}

	var result *types.Command

	err = c.Find(bson.M{"command_id": command_id}).One(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *MongoDatabase) DeleteCommand(ref *string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("commands")

	command_id, err := strconv.ParseInt(*ref, 0, 0)
	if nil != err {
		return err
	}
	err = c.Remove(bson.M{"command_id": command_id})
	if err != nil {
		return err
	}

	return nil
}

func (d *MongoDatabase) AddDownloadLink(link *types.DownloadLink) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("download_links")

	err = c.Insert(link)
	if err != nil {
		return errors.New("Register error")
	}

	d.Log(DEBUG, fmt.Sprintf("%s: %s", "Saving download link", link))
	return nil
}

func (d *MongoDatabase) GetDownloadLink(ref string) (link *types.DownloadLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	c := d.session.DB("mongo").C("download_links")

	var result *types.DownloadLink
	err = c.Find(bson.M{"link": ref}).One(&result)
	if (err != nil) || (result == nil) {
		d.Log(ERROR, fmt.Sprintf("%s, %s", "Couldn't find download link", ref))
		return nil, errors.New(fmt.Sprintf("%s: %s", "Couldn't find this downloadLink", ref))
	}

	return result, nil
}

func (d *MongoDatabase) AddAccount(account *types.Account) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("accounts")
	//Check if same user already exists
	result, err := c.Find(bson.M{"$or": []bson.M{{"login": account.Login}, {"email": account.Email}}}).Count()
	if (err != nil) || (result != 0) {
		return errors.New("Account already exists")
	}
	index, err := c.Find(nil).Count()

	account.Id = strconv.Itoa(index + 1)
	err = c.Insert(account)
	if err != nil {
		return errors.New("Register error")
	}
	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Account", account))

	return nil
}

func (d *MongoDatabase) GetAccount(authType string, ref string) (account *types.Account, id string, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getAccount(authType, ref)
}

func (d *MongoDatabase) getAccount(authType string, ref string) (account *types.Account, id string, err error) {
	c := d.session.DB("mongo").C("accounts")

	var result *types.Account
	err = c.Find(bson.M{"$or": []bson.M{{"login": ref}, {"email": ref}}}).One(&result)
	if err != nil {
		message := fmt.Sprintf("GetAccount: Couldn't find the desired account %s:%s", authType, ref)
		d.Log(ERROR, message)
		fmt.Println(err)
		return nil, "", errors.New(message)
	}

	if 0 == len(authType) {
		return result, result.Id, nil
	}
	_, found := result.Auths[authType]
	if !found {
		message := fmt.Sprintf("Couldn't find this kind of authentication %s for %s", authType, ref)
		d.Log(ERROR, message)
		return nil, "", errors.New(message)
	}

	return result, result.Id, nil
}

func (d *MongoDatabase) UpdateAccount(id string, account *types.Account) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("accounts")
	account_id, err := strconv.ParseInt(id, 0, 0)
	if nil != err {
		return err
	}

	err = c.Update(bson.M{"id": account_id}, account)
	if nil != err {
		return err
	}
	return nil
}

func (d *MongoDatabase) GetUserAccount(id string) (account *types.Account, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	c := d.session.DB("mongo").C("accounts")

	var result *types.Account
	err = c.Find(bson.M{"id": id}).One(&result)
	if nil != err {
		message := fmt.Sprintf("GetUserAccount: Couldn't find the desired account %s", id)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}
	return result, nil
}

func (d *MongoDatabase) ListAccounts(searchDict map[string]string) (accounts []*types.Account, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	c := d.session.DB("mongo").C("accounts")
	var results []*types.Account

	err = c.Find(nil).All(&results)

	if 0 == len(searchDict) {
		d.Log(DEBUG, "No search parameters")
		return results, nil
	}

	d.Log(DEBUG, fmt.Sprintf("We had some search parameters ", searchDict))
	accounts = make([]*types.Account, len(results))
	i := 0
	for _, account := range results {
		for k, v := range searchDict {
			switch k {
			case "login":
				if strings.Contains(account.Login, v) {
					accounts[i] = account
					i += 1
					break
				}
			case "email":
				if strings.Contains(account.Email, v) {
					accounts[i] = account
					i += 1
					break
				}
			case "id":
				if strings.Contains(account.Id, v) {
					accounts[i] = account
					i += 1
					break
				}
			}
		}
	}
	return accounts[0:i], nil
}

func (d *MongoDatabase) StoreSession(session *types.Session) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("sessions")
	count, err := c.Find(bson.M{"user_id": session.UserId}).Count()
	if err != nil {
		return errors.New("Register error")
	}
	if count != 0 {
		return nil
	}
	err = c.Insert(session)
	if err != nil {
		return errors.New("Register error")
	}
	return nil
}

func (d *MongoDatabase) GetSession(ref string) (session *types.Session, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	c := d.session.DB("mongo").C("sessions")
	var result *types.Session
	err = c.Find(bson.M{"authentication_header": ref}).One(&result)
	if err != nil {
		return nil, errors.New("Couldn't find session")
	}
	return result, nil
}

func (d *MongoDatabase) RemoveSession(ref string) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("sessions")
	err = c.Remove(bson.M{"authentication_header": ref})
	if err != nil {
		return err
	}
	return nil
}

func (d *MongoDatabase) SaveShareLink(shareLink *types.ShareLink) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	//TODO check if there is already a sharelink with this name and user
	c := d.session.DB("mongo").C("shares_link")
	shareLink.Id = *shareLink.ShareLink.Key
	err = c.Insert(shareLink)
	if err != nil {
		return errors.New("Register SaveShareLink error")
	}
	return nil
}

func (d *MongoDatabase) GetShareLink(key string) (shareLink *types.ShareLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	c := d.session.DB("mongo").C("shares_link")
	var result *types.ShareLink
	err = c.Find(bson.M{"id": key}).One(&result)
	if err != nil {
		message := fmt.Sprintf("Couldn't find share link %s", key)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}
	return result, nil
}

func (d *MongoDatabase) RemoveShareLink(key string) (err error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("shares_link")
	err = c.Remove(bson.M{"id": key})
	if err != nil {
		return err
	}
	return nil
}

func (d *MongoDatabase) ListShareLinks(user string) (shareLinks []*types.ShareLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	c := d.session.DB("mongo").C("shares_link")
	var results []*types.ShareLink
	err = c.Find(bson.M{"user": user}).All(&results)
	if err != nil {
		message := fmt.Sprintf("Couldn't find share link for user %s", user)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}
	return shareLinks, err
}

func (d *MongoDatabase) GetShareLinksFromPath(path string, user string) (shareLink []*types.ShareLink, err error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getShareLinksFromPath(path, user)
}

func (d *MongoDatabase) getShareLinksFromPath(path string, user string) (shareLinks []*types.ShareLink, err error) {
	c := d.session.DB("mongo").C("shares_link")
	var results []*types.ShareLink
	err = c.Find(nil).All(&results)
	// TODO optimize it with and expression, pb: access to shareLink.ShareLink.Path in find request
	if err != nil {
		message := fmt.Sprintf("Couldn't find share link for user %s path %s", user, path)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}
	shareLinks = make([]*types.ShareLink, 0, 10)
	var count = 0
	for _, shareLink := range results {
		if *shareLink.ShareLink.Path == path && shareLink.User == user {
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

func (d *MongoDatabase) GetAccess(user *string, path string) (api.AccessType, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.getAccess(user, path)
}

func (d *MongoDatabase) getAccess(user *string, path string) (api.AccessType, error) {
	c := d.session.DB("mongo").C("accesses")
	var access pathAccesses
	var user_id string
	if nil == user {
		user_id = ""
	} else {
		user_id = *user
	}

	message := fmt.Sprintf("GetAccess: %s path %s", user_id, path)
	d.Log(ERROR, message)
	err := c.Find(bson.M{"user_id": user_id}).One(&access)
	if err != nil {
		// If user is not Nil then look for the public accesses
		if nil == user {
			return api.NONE, nil
		}
		err = c.Find(bson.M{"user_id": ""}).One(&access)
		if err != nil {
			return api.NONE, nil
		}
	}

	// Now check for the path
	splittedPath := strings.Split(path, "/")
	finalAccessType := api.NONE
	for i := len(splittedPath); i > 0; i-- {
		accessType, found := access.Accesses[strings.Join(splittedPath[:i], "/")]
		tools.LOG_DEBUG.Println("Looking for access to ", strings.Join(splittedPath[:i], "/"))
		if found {
			finalAccessType = accessType
			break
		}
	}
	if api.NONE == finalAccessType && nil != user {
		return d.getAccess(nil, path)
	}

	return finalAccessType, nil
}

func (d *MongoDatabase) SetAccess(user *string, path string, access api.AccessType) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	var user_mail string
	if nil == user {
		user_mail = ""
	} else {
		user_mail = *user
	}
	// Convert mail in to user_id
	_, user_id, err := d.getAccount("", user_mail)
	if err != nil {
		message := fmt.Sprintf("SetAccess : user_id not found %s path %s", user_mail, path)
		d.Log(ERROR, message)
		return nil
	}

	c := d.session.DB("mongo").C("accesses")
	var path_access pathAccesses
	err = c.Find(bson.M{"user_id": user_id}).One(&path_access)
	message := fmt.Sprintf("SetAccess %s path %s", user_mail, path)
	d.Log(ERROR, message)
	if err != nil {
		//Create accessPath dict
		message = fmt.Sprintf("SetAccess create with %s path %s", user_mail, path)
		d.Log(ERROR, message)
		accesses := new(pathAccesses)
		accesses.Accesses = make(map[string]api.AccessType)
		accesses.UserId = user_id
		err = c.Insert(accesses)
		if err != nil {
			message := fmt.Sprintf("Couldn't insert access %s path %s", user_mail, path)
			d.Log(ERROR, message)
			return errors.New(message)
		}
		return nil
	}
	message = fmt.Sprintf("SetAccess update with %s path %s", user_mail, path)
	d.Log(ERROR, message)
	path_access.Accesses[path] = access
	err = c.Update(bson.M{"user_mail": user_mail}, path_access)
	if err != nil {
		message := fmt.Sprintf("Couldn't update access %s path %s", user_mail, path)
		d.Log(ERROR, message)
		return errors.New(message)
	}
	return nil
}

func (d *MongoDatabase) ClearAccesses() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	c := d.session.DB("mongo").C("accesses")
	err := c.DropCollection()
	if err != nil {
		// No entry in the database
		return nil
	}
	return nil
}
