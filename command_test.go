package lilraft

var (
	sNum uint64 = 0
)

type testCommand struct {
	serialNum uint64
}

func (t *testCommand) Name() string {
	return "test"
}

func (t *testCommand) SerialNum() uint64 {
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
