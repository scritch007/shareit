//shareit package aims at browsing files and sharing them with others
package shareit

//CommandHandler is used to keep information about issued commands
type CommandHandler struct {
	config        *Configuration
	commandsList  []*Command
	commandIndex  int
	downloadLinks map[string]string
}

//Configuration Structure
type Configuration struct {
	RootPrefix string
	PrivateKey string
	StaticPath string
	WebPort    string
}