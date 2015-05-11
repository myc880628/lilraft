package lilraft

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// commandType stores client's command reference
var commandType = make(map[string]Command)

// Command is a interface for client to implement and send to server
type Command interface {
	SerialNum() int64          // server use this to distinguish command, in case to execute twice
	Apply(context interface{}) // server will use Apply() to run the command
	Name() string              // The command's name
}

func newCommand(name string, commandData []byte) (Command, error) {
	command := commandType[name]
	if command == nil {
		return nil, fmt.Errorf("command not registered")
	}
	logger.Println(commandData)
	if err := gob.NewDecoder(bytes.NewReader(commandData)).Decode(command); err != nil {
		return nil, err
	}
	return command, nil
}
