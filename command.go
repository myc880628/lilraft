package lilraft

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// commandType stores client's command reference
var commandType map[string]Command

// Command is a interface for client to implement and send to server
type Command interface {
	SerialNum() int64 // server use this to distinguish command, in case to execute twice
	Apply()           // server will use Apply() to run the command
	Name() string     // The command's name
}

func newCommand(name string, commandData []byte) (Command, error) {
	command := commandType[name]
	if command == nil {
		return nil, fmt.Errorf("command not registered")
	}

	if err := json.NewDecoder(bytes.NewReader(commandData)).Decode(command); err != nil {
		return nil, err
	}
	return command, nil
}
