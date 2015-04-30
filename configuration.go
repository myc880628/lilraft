package lilraft

const (
	c_old = iota
	c_old_new
)

type configuration struct {
	mutexState // configuration can have C_old state or C_old_new state
	c_OldNode  nodeMap
	c_NewNode  nodeMap
}

// TODO: fill the pass function
func (c *configuration) pass() bool {
	return true
}
