package database

import (
	"encoding/json"
	"errors"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/database/dummy"
	"github.com/scritch007/shareit/database/mongo"
	//"github.com/scritch007/shareit/database/sqlite"
	"github.com/scritch007/shareit/types"
)

func NewDatabase(name string, config *json.RawMessage, debug bool) (types.DatabaseInterface, error) {
	tools.LOG_DEBUG.Println("Creating new instance of database", name)
	var newDatabase types.DatabaseInterface
	var err error
	switch name {
	case dummy.Name:
		newDatabase, err = dummy.NewDummyDatabase(config)
	case mongo.Name:
		newDatabase, err = mongo.NewMongoDatase(config)
	/*case sqlite.Name:
	newDatabase, err = sqlite.NewSqliteDatase(config, debug)*/
	default:
		err = errors.New("Unknown authentication method " + name)
		newDatabase = nil
	}
	return newDatabase, err
}
