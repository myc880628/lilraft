package lilraft

import (
	"testing"
)

// type Command interface {
// 	SerialNum() uint64 // server use this to distinguish command, in case to execute twice
// 	Apply()            // server will use Apply() to run the command
// 	Name() string      // The command's name
// }

type testCommand struct {
}

func (t *testCommand) Name() string {
	return "name"
}

func (t *testCommand) SerialNum() uint64 {
	return 1
}

func (t *testCommand) Apply() {
	println("hello server")
}

func TestCreateLogEntry(t *testing.T) {
	_, err := newLogEntry(2, 2, &testCommand{})
	if err != nil {
		t.Error(err)
	}
}
