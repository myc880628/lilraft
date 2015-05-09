package lilraft

type testCommand struct {
	Serial int64 // Use capital to be encoded and decoded by json package
	Val    int
}

func (t *testCommand) Name() string {
	return "test"
}

func (t *testCommand) SerialNum() int64 {
	return t.Serial
}

func (t *testCommand) Apply(context interface{}) {
	sl := context.(*[]int)
	logger.Printf("array appended %d\n", t.Val)
	*sl = append(*sl, t.Val)
}

func newTestCommand(sNum int64, val int) *testCommand {
	return &testCommand{
		Serial: sNum,
		Val:    val,
	}
}
