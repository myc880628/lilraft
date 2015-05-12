package lilraft

type testCommand struct {
	Val int
}

func (t *testCommand) Name() string {
	return "test"
}

func (t *testCommand) Apply(context interface{}) {
	sl := context.(*[]int)
	logger.Printf("array appended %d\n", t.Val)
	*sl = append(*sl, t.Val)
}

func newTestCommand(val int) *testCommand {
	return &testCommand{
		Val: val,
	}
}
