package lilraft

var commandType map[string]Command

// Command is a interface for client to implement and send to server
type Command interface {
	Index() uint64 //Index should return an index of command, in case that a command is executed twice
	Apply()        // server will use Apply() to run the command
	Name() string  // The command's name
}
