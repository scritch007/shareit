package database
import (
	"errors"
	"encoding/json"
	"github.com/scritch007/shareit/types"
	"github.com/scritch007/shareit/database/dummy"
)

func NewDatabase(name string, config *json.RawMessage) (types.DatabaseInterface, error) {
	types.LOG_DEBUG.Println("Creating new instance of database %s", name)
	var newDatabase types.DatabaseInterface
	var err error
	switch name{
	case dummy.Name:
		newDatabase, err = dummy.NewDummyDatabase(config)
	default:
		err = errors.New("Unknown authentication method " + name)
		newDatabase = nil
	}
	return newDatabase, err
}