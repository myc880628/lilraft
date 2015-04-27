package lilraft_test

import "fmt"

type TestCommand struct {
}

func (T *TestCommand) Apply() {
	fmt.Println("Test command Applying")
}

func (T *TestCommand) Name() string {
	return "Test"
}
