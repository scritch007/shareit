package dummy

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"errors"
	"github.com/scritch007/shareit/types"
	//"github.com/scritch007/shareit/database"
)

const(
	Name string = "DummyDb"
)

type DummyDatabase struct{
	DbFolder string `json:"db_folder"`
	commandsList  []*types.Command
	commandIndex  int
	downloadLinks map[string]*types.DownloadLink
}

func NewDummyDatabase(config *json.RawMessage)(d *DummyDatabase, err error){
	d = new(DummyDatabase)
	d.commandsList = make([]*types.Command, 10)
	d.commandIndex = 0
	d.downloadLinks = make(map[string]*types.DownloadLink)
	if err = json.Unmarshal(*config, d); nil != err{
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

func (d *DummyDatabase)Name() string{
	return Name
}
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)
func (d *DummyDatabase)Log(level LogLevel, message string){
	switch(level){
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
func (d *DummyDatabase)AddCommand(command *types.Command) (ref string, err error){
	fmt.Println(command)
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
	d.Log(DEBUG, fmt.Sprintf( "%s : %s", "Saved new Command", command))
	return ref, nil
}
func (d *DummyDatabase)ListCommands(offset int, limit int, search_parameters *types.CommandsSearchParameters) ([]*types.Command, int, error){
	return d.commandsList[0:d.commandIndex], d.commandIndex, nil
}
func (d *DummyDatabase)GetCommand(ref string)(command *types.Command, err error){
	command_id, err := strconv.ParseInt(ref, 0, 0)
	if nil != err{
		return nil, err
	}
	command = d.commandsList[command_id]
	return command, nil
}
func (d *DummyDatabase)DeleteCommand(ref *string) error{
	return nil
}
func (d *DummyDatabase)AddDownloadLink(link *types.DownloadLink)(err error){
	d.Log(DEBUG, fmt.Sprintf("%s: %s", "Saving download link", link))
	d.downloadLinks[link.Link] = link
	return nil
}
func (d *DummyDatabase)GetDownloadLink(ref string)(link *types.DownloadLink, err error){
	res, found := d.downloadLinks[ref]
	if !found{
		d.Log(ERROR, fmt.Sprintf("%s, %s", "Couldn't find download link", ref))
		return nil, errors.New(fmt.Sprintf("%s: %s", "Couldn't find this downloadLink", ref))
	}
	return res, nil
}