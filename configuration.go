package lilraft

var cOldNewStr = "Cold,new"
var cNewStr = "Cnew"

// implement joint consensus
type configuration struct {
	mutexState // configuration can have C_old state or C_old_new state
	c_OldNode  nodeMap
	c_NewNode  nodeMap
}

// TODO: add more procedures
func NewConfig(nodes ...Node) (conf *configuration) {
	conf = &configuration{
		c_OldNode: makeNodeMap(nodes...),
	}
	conf.setState(c_old)
	return
}
