#lilraft - a [raft paper](https://raftconsensus.github.io/) implementation in Go

###Usage:

#####Create a new server and start:

```go
var array []int // array is the custom context for clients
config := lilraft.NewConfig(
        lilraft.NewHTTPNode(1, "http://127.0.0.1:8787"),
        lilraft.NewHTTPNode(2, "http://127.0.0.1:8788"),
        lilraft.NewHTTPNode(3, "http://127.0.0.1:8789"),
)
// Set server id, custom contex, configuration and server path
s := lilraft.NewServer(1, &array, config, "server/")
s.SetHTTPTransport(http.DefaultServeMux, 8787)
s.Start(false) // set recoverContex flag to false
...
s.Stop()
```

#####Execute a command
To execute a command, you need to implement ```Command``` interface, which contains ```Apply(context interface{})```
and ```Name() string```, when client execute a command, server will execute the ```Apply``` method
```go
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
...
s.RegisterCommand(&testCommand{}) // Register a command so other server nodes can decode data package containing command
...
if err := s.Exec(newTestCommand(2046)); err != nil {
        fmt.Println(err.Error())
}
```

####Recover custom contex from log
Simply set recoverContex flag to true when start a server.
Note: Only committed log entries will be wirtten to file
```go
s.Start(true)
```
####Configuration change
Raft protocol use two phases to change configuration. For more details, please check out the raft paper.
```go
s.SetConfig(
        lilraft.NewHTTPNode(1, "http://127.0.0.1:8787"),
        lilraft.NewHTTPNode(2, "http://127.0.0.1:8788"),
        lilraft.NewHTTPNode(3, "http://127.0.0.1:8789"),
        lilraft.NewHTTPNode(4, "http://127.0.0.1:8790"),
        lilraft.NewHTTPNode(5, "http://127.0.0.1:8791"),
)
```

For more concrete examples, check out the example folder
