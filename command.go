package lilraft

var commandType map[string]Command

// Command is a interface for client to implement and send to server
type Command interface {
	SerialNum()   // server use this to distinguish command, in case to execute twice
	Apply()       // server will use Apply() to run the command
	Name() string // The command's name
}
