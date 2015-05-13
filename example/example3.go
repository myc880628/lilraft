package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/lilwulin/lilraft"
)

type testCommand struct {
	Val int
}

func (t *testCommand) Name() string {
	return "test"
}

func (t *testCommand) Apply(context interface{}) {
	sl := context.(*[]int)
	*sl = append(*sl, t.Val)
}

func newTestCommand(val int) *testCommand {
	return &testCommand{
		Val: val,
	}
}

func main() {
	var array []int
	config := lilraft.NewConfig(
		lilraft.NewHTTPNode(1, "http://127.0.0.1:8787"),
	)
	s := lilraft.NewServer(1, &array, config, "server/")
	s.SetHTTPTransport(http.DefaultServeMux, 8787)
	s.RegisterCommand(&testCommand{})
	s.Start(true)
	time.Sleep(500 * time.Millisecond)
	fmt.Println(array)
	s.Stop()
}
