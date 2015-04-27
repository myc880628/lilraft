package lilraft_test

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Command interface {
	Apply()
	Name() string
}

type TestCommand struct {
	MyName string
}

func (T *TestCommand) Apply() {
	println("I am applying")
}

func (T *TestCommand) SetName() {
	T.MyName = "Set!"
}

func (T *TestCommand) Name() string {
	return "Test"
}

var mc = make(map[string]Command)

func Register(command Command) {
	mc[command.Name()] = command
}

func main() {
	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)
	dec := gob.NewDecoder(m)
	var tcomm = TestCommand{MyName: "hei"}
	if err := enc.Encode(tcomm); err != nil {
		log.Fatal("encode error:", err)
	}

	Register(&TestCommand{})

	var comm Command = mc["Test"]
	if err := dec.Decode(comm); err != nil {
		log.Fatal("decode erorr:", err)
	}

	println(comm.Name())
	println(comm.(*TestCommand).MyName)
}
