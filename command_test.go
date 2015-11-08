package lilraft

type testCommand struct {
	Val int
}

func (t *testCommand) Name() string {
	return "test"
}

func (t *testCommand) Apply(context interface{}) (interface{}, error) {
	sl := context.(*[]int)
	logger.Printf("array appended %d\n", t.Val)
	*sl = append(*sl, t.Val)
	return nil, nil
}

func newTestCommand(val int) *testCommand {
	return &testCommand{
		Val: val,
	}
}

type testCommand2 struct {
}

func (t *testCommand2) Name() string {
	return "test2"
}

func (t *testCommand2) Apply(context interface{}) (interface{}, error) {
	sl := context.(*[]int)
	return (*sl)[0], nil
}

func newTestCommand2() *testCommand2 {
	return &testCommand2{}
}
