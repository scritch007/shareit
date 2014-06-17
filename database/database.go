package database

type DatabaseInterface interface{
	Name() string
	
}

func NewDatabase(name string, config_path string) error {

}