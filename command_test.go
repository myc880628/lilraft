package lilraft

var (
	sNum int64 = 0
)

type testCommand struct {
	serialNum int64
}

func (t *testCommand) Name() string {
	return "test"
}

func (t *testCommand) SerialNum() int64 {
	return t.serialNum
}

func (t *testCommand) Apply() {
	println("hello lilraft")
}

func newTestCommand() *testCommand {
	sNum++
	return &testCommand{
		serialNum: sNum,
	}
}
