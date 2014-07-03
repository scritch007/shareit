package dummy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/scritch007/shareit/types"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	//"github.com/scritch007/shareit/database"
)

const (
	Name string = "DummyDb"
)

type DummyDatabase struct {
	DbFolder       string `json:"db_folder"`
	commandsList   []*types.Command
	commandIndex   int
	downloadLinks  map[string]*types.DownloadLink
	accounts       []*types.Account
	accountsId     int
	accountsDBPath string
	sessionMap     map[string]*types.Session
	shareLinkMap   map[string]*types.ShareLink
	shareLinkPath  string
}

func NewDummyDatabase(config *json.RawMessage) (d *DummyDatabase, err error) {
	d = new(DummyDatabase)
	d.commandsList = make([]*types.Command, 10)
	d.commandIndex = 0
	d.downloadLinks = make(map[string]*types.DownloadLink)
	d.sessionMap = make(map[string]*types.Session)
	if err = json.Unmarshal(*config, d); nil != err {
		return nil, err
	}
	//Prepare the folder
	if _, err := os.Stat(d.DbFolder); err != nil {
		if os.IsNotExist(err) {
			types.LOG_ERROR.Println("Error: the path %s, doesn't exist", d.DbFolder)
		} else {
			types.LOG_ERROR.Println("Error: Something went wrong when accessing to %s, %v", d.DbFolder, err)
		}
		return nil, err
	}
	d.accountsDBPath = path.Join(d.DbFolder, "accounts.json")
	if _, err := os.Stat(d.accountsDBPath); err != nil {
		var fo *os.File
		if os.IsNotExist(err) {
			fo, err = os.Create(d.accountsDBPath)
			if nil != err {
				return nil, err
			}
			defer fo.Close()
		} else {
			return nil, err
		}
		d.accounts = make([]*types.Account, 10)
		d.accountsId = 0
		d.saveDb()
	} else {
		file, err := ioutil.ReadFile(d.accountsDBPath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(file, &d.accounts)
		if nil != err {
			return nil, err
		}
		d.accountsId = len(d.accounts)
		if 0 == d.accountsId {
			//Special case we need to allocate some memory anyway
			d.accounts = make([]*types.Account, 10)
		}
	}
	d.shareLinkPath = path.Join(d.DbFolder, "share_links.json")
	if _, err := os.Stat(d.shareLinkPath); err != nil {
		var fo *os.File
		if os.IsNotExist(err) {
			fo, err = os.Create(d.shareLinkPath)
			if nil != err {
				return nil, err
			}
			defer fo.Close()
		} else {
			return nil, err
		}
		d.shareLinkMap = make(map[string]*types.ShareLink)
		d.saveDb()
	} else {
		file, err := ioutil.ReadFile(d.shareLinkPath)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(file, &d.shareLinkMap)
		if nil != err {
			return nil, err
		}
	}
	return d, nil
}

func (d *DummyDatabase) Name() string {
	return Name
}

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

func (d *DummyDatabase) Log(level LogLevel, message string) {
	switch level {
	case DEBUG:
		types.LOG_DEBUG.Println("DummyDb: ", message)
	case INFO:
		types.LOG_INFO.Println("DummyDb: ", message)
	case WARNING:
		types.LOG_WARNING.Println("DummyDb: ", message)
	case ERROR:
		types.LOG_ERROR.Println("DummyDb: ", message)
	}
}
func (d *DummyDatabase) SaveCommand(command *types.Command) (err error) {
	if 0 == len(command.CommandId) {
		d.commandsList[d.commandIndex] = command
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
	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Command", command))
	return nil
}
func (d *DummyDatabase) ListCommands(user *string, offset int, limit int, search_parameters *types.CommandsSearchParameters) ([]*types.Command, int, error) {
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
func (d *DummyDatabase) GetCommand(ref string) (command *types.Command, err error) {
	command_id, err := strconv.ParseInt(ref, 0, 0)
	if nil != err {
		return nil, err
	}
	command = d.commandsList[command_id]
	return command, nil
}

func (d *DummyDatabase) DeleteCommand(ref *string) error {
	return nil
}
func (d *DummyDatabase) AddDownloadLink(link *types.DownloadLink) (err error) {
	d.Log(DEBUG, fmt.Sprintf("%s: %s", "Saving download link", link))
	d.downloadLinks[link.Link] = link
	return nil
}
func (d *DummyDatabase) GetDownloadLink(ref string) (link *types.DownloadLink, err error) {
	res, found := d.downloadLinks[ref]
	if !found {
		d.Log(ERROR, fmt.Sprintf("%s, %s", "Couldn't find download link", ref))
		return nil, errors.New(fmt.Sprintf("%s: %s", "Couldn't find this downloadLink", ref))
	}
	return res, nil
}

func (d *DummyDatabase) AddAccount(account *types.Account) (err error) {

	i := 0
	var item *types.Account
	//Iter once to check if same user already exists
	for i < d.accountsId {
		item = d.accounts[i]
		i += 1
		if (item.Login == account.Login) || (item.Email == account.Email) {
			return errors.New("Account already exists")
		}
	}

	if len(d.accounts) == d.accountsId {
		new_list := make([]*types.Account, len(d.accounts)*2)
		for i, comm := range d.accounts {
			new_list[i] = comm
		}
		d.accounts = new_list
	}
	//Todo Check that no other account has the same (Id, authType)
	d.accounts[d.accountsId] = account
	account.Id = account.Email
	d.accountsId += 1

	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Account", account))

	return d.saveDb()
}
func (d *DummyDatabase) GetAccount(authType string, ref string) (account *types.Account, err error) {
	for _, elem := range d.accounts[0:d.accountsId] {
		if (ref == elem.Email) || (ref == elem.Login) {
			_, found := elem.Auths[authType]
			if !found {
				message := fmt.Sprintf("Couldn't find this kind of authentication %s for %s", authType, ref)
				d.Log(ERROR, message)
				return nil, errors.New(message)
			}
			return elem, nil
		}
	}
	message := fmt.Sprintf("Couldn't find the desired account %s:%s", authType, ref)
	d.Log(ERROR, message)
	return nil, errors.New(message)
}
func (d *DummyDatabase) GetUserAccount(id string) (account *types.Account, err error) {
	for _, elem := range d.accounts[0:d.accountsId] {
		d.Log(DEBUG, fmt.Sprintf("Looking for %s comparing with %s", id, elem.Id))
		if id == elem.Id {
			return elem, nil
		}
	}
	message := fmt.Sprintf("Couldn't find the desired account %s", id)
	d.Log(ERROR, message)
	return nil, errors.New(message)
}

func (d *DummyDatabase) ListAccounts(searchDict map[string]string) (accounts []*types.Account, err error) {
	if 0 == len(searchDict) {
		d.Log(DEBUG, "No search parameters")
		return d.accounts[0:d.accountsId], nil
	}
	d.Log(DEBUG, fmt.Sprintf("We had some search parameters ", searchDict))

	accounts = make([]*types.Account, d.accountsId)

	i := 0
	for _, account := range d.accounts[0:d.accountsId] {
		for k, v := range searchDict {
			switch k {
			case "login":
				if strings.Contains(account.Login, v) {
					accounts[i] = account
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

func (d *DummyDatabase) StoreSession(session *types.Session) (err error) {
	d.sessionMap[session.AuthenticationHeader] = session
	return nil
}
func (d *DummyDatabase) GetSession(ref string) (session *types.Session, err error) {
	session, found := d.sessionMap[ref]
	if !found {
		return nil, errors.New("Couldn't find session")
	}
	return session, nil
}

func (d *DummyDatabase) RemoveSession(ref string) (err error) {
	delete(d.sessionMap, ref)
	return nil
}

func (d *DummyDatabase) SaveShareLink(shareLink *types.ShareLink) (err error) {
	//TODO check if there is already a sharelink with this name and user
	_, err = d.GetShareLinkFromPath(*shareLink.Path, shareLink.User)
	if nil == err {
		err = errors.New("Share link already exists")
		return err
	}
	d.shareLinkMap[*shareLink.Key] = shareLink
	return d.saveDb()
}
func (d *DummyDatabase) GetShareLink(key string) (shareLink *types.ShareLink, err error) {
	shareLink, found := d.shareLinkMap[key]
	if !found {
		message := fmt.Sprintf("Couldn't find share link %s", key)
		d.Log(ERROR, message)
		return nil, errors.New(message)
	}
	return shareLink, nil
}
func (d *DummyDatabase) RemoveShareLink(key string) (err error) {
	delete(d.shareLinkMap, key)
	return d.saveDb()
}
func (d *DummyDatabase) ListShareLinks(user string) (shareLinks []*types.ShareLink, err error) {
	shareLinks = make([]*types.ShareLink, 30)
	currentId := 0
	for _, shareLink := range d.shareLinkMap {
		if shareLink.User == user {
			shareLinks[currentId] = shareLink
			currentId += 1
			if currentId == len(shareLinks) {
				new_list := make([]*types.ShareLink, len(shareLinks)*2)
				for i, comm := range shareLinks {
					new_list[i] = comm
				}
				shareLinks = new_list
			}
		}
	}
	return shareLinks, err
}
func (d *DummyDatabase) GetShareLinkFromPath(path string, user string) (shareLink *types.ShareLink, err error) {
	for _, shareLink = range d.shareLinkMap {
		if *shareLink.Path == path && shareLink.User == user {
			return shareLink, nil
		}
	}
	return nil, errors.New("Couldn't find shareLink")
}

func (d *DummyDatabase) saveDb() error {
	serialized, err := json.Marshal(d.shareLinkMap)
	if nil != err {
		d.Log(ERROR, "Couldn't serialize share links...")
		return err
	}

	var fo *os.File
	fo, err = os.OpenFile(d.shareLinkPath, os.O_WRONLY, os.ModePerm)
	if nil != err {
		return err
	}
	defer fo.Close()
	nbWriten, err := fo.Write(serialized)
	if nbWriten != len(serialized) {
		d.Log(ERROR, "Couldn't write serialized object")
		return errors.New("Couldn't write serialized object")
	}

	serialized, err = json.Marshal(d.accounts[0:d.accountsId])
	if nil != err {
		d.Log(ERROR, "Couldn't serialize accounts list...")
		return err
	}

	fo, err = os.OpenFile(d.accountsDBPath, os.O_WRONLY, os.ModePerm)
	if nil != err {
		return err
	}
	defer fo.Close()
	nbWriten, err = fo.Write(serialized)
	if nbWriten != len(serialized) {
		d.Log(ERROR, "Couldn't write serialized object")
		return errors.New("Couldn't write serialized object")
	}

	return nil
}
