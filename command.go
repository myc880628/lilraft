package lilraft

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// commandType stores client's command reference
var commandType = make(map[string]Command)
var emptyCommandName = "empty"

// Command is a interface for client to implement and send to server
type Command interface {
	Apply(context interface{}) (interface{}, error) // server will use Apply() to run the command
	Name() string                                   // The command's name
}

func newCommand(name string, commandData []byte) (Command, error) {
	command := commandType[name]
	if command == nil && name != cOldNewStr && name != cNewStr {
		return nil, fmt.Errorf("command not registered")
	}
	if err := json.NewDecoder(bytes.NewReader(commandData)).Decode(command); err != nil {
		return nil, err
	}
	return command, nil
}

type emptyCommand struct{}

func (ec *emptyCommand) Apply(context interface{}) {}

func (ec *emptyCommand) Name() string {
	return emptyCommandName
}
