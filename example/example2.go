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

func (t *testCommand) Apply(context interface{}) (interface{}, error) {
	sl := context.(*[]int)
	*sl = append(*sl, t.Val)
	return nil, nil
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
	s.Start(false)
	time.Sleep(400 * time.Millisecond)
	if err := s.Exec(newTestCommand(2046)); err != nil {
		fmt.Println(err.Error())
	}
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	fmt.Println(array)
}
