package dummy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/scritch007/shareit/types"
	"os"
	"path"
	"strconv"
	//"github.com/scritch007/shareit/database"
)

const (
	Name string = "DummyDb"
)

type DummyDatabase struct {
	DbFolder      string `json:"db_folder"`
	commandsList  []*types.Command
	commandIndex  int
	downloadLinks map[string]*types.DownloadLink
	accounts      []*types.Account
	accountsId    int
}

func NewDummyDatabase(config *json.RawMessage) (d *DummyDatabase, err error) {
	d = new(DummyDatabase)
	d.commandsList = make([]*types.Command, 10)
	d.commandIndex = 0
	d.downloadLinks = make(map[string]*types.DownloadLink)
	d.accountsId = 0
	d.accounts = make([]*types.Account, 10)
	if err = json.Unmarshal(*config, d); nil != err {
		return nil, err
	}
	//Prepare the folder
	if _, err := os.Stat(d.DbFolder); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Error: the path %s, doesn't exist", d.DbFolder)
		} else {
			fmt.Println("Error: Something went wrong when accessing to %s, %v", d.DbFolder, err)
		}
		return nil, err
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
func (d *DummyDatabase) AddCommand(command *types.Command) (ref string, err error) {
	d.commandsList[d.commandIndex] = command
	ref = strconv.Itoa(d.commandIndex)
	d.commandIndex += 1
	if len(d.commandsList) == d.commandIndex {
		new_list := make([]*types.Command, len(d.commandsList)*2)
		for i, comm := range d.commandsList {
			new_list[i] = comm
		}
		d.commandsList = new_list
	}
	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Command", command))
	return ref, nil
}
func (d *DummyDatabase) ListCommands(offset int, limit int, search_parameters *types.CommandsSearchParameters) ([]*types.Command, int, error) {
	return d.commandsList[0:d.commandIndex], d.commandIndex, nil
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

	//Iter once to check if same user already exists
	for _, item := range d.accounts {
		if (item.Login == account.Login) || (item.Email == account.Email) {
			return errors.New("Account already exists")
		}
	}

	//Todo Check that no other account has the same (Id, authType)
	d.accounts[d.accountsId] = account
	d.accountsId += 1
	if len(d.accounts) == d.accountsId {
		new_list := make([]*types.Account, len(d.accounts)*2)
		for i, comm := range d.accounts {
			new_list[i] = comm
		}
		d.accounts = new_list
	}
	d.Log(DEBUG, fmt.Sprintf("%s : %s", "Saved new Account", account))

	serialized, err := json.Marshal(d.accounts[0:d.accountsId])
	if nil != err {
		d.Log(ERROR, "Couldn't serialize accounts list...")
		return err
	}
	accountsDBPath := path.Join(d.DbFolder, "accounts.json")
	var fo *os.File
	if _, err := os.Stat(accountsDBPath); err != nil {
		if os.IsNotExist(err) {
			fo, err = os.Create(accountsDBPath)
			if nil != err {
				return err
			}
		} else {
			return err
		}
	} else {
		fo, err = os.OpenFile(accountsDBPath, os.O_WRONLY, os.ModePerm)
		if nil != err {
			return err
		}
	}
	nbWriten, err := fo.Write(serialized)
	if nbWriten != len(serialized) {
		d.Log(ERROR, "Couldn't write serialized object")
		return errors.New("Couldn't write serialized object")
	}
	err = fo.Close()
	if nil != err {
		d.Log(ERROR, "Failed to close the file")
		return err
	}

	return nil
}
func (d *DummyDatabase) GetAccount(authType string, ref string) (account *types.Account, err error) {
	for _, elem := range d.accounts {
		if (authType == elem.AuthType) && (ref == elem.Id) {
			return elem, nil
		}
	}
	message := fmt.Sprintf("Couldn't find the desired account %s:%s", authType, ref)
	d.Log(ERROR, message)
	return nil, errors.New(message)
}
