package authentication


type DummyAuth struct{

}

func (d *DummyAuth)Name() string{
	return "DummyAuth"
}
func (d *DummyAuth)AddRoutes(r *Router) error{
	return nil
}
func (d *DummyAuth)ParseConfig(configPath string) error{
	return nil
}