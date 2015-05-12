package lilraft

var cOldNewStr = "Cold,new"
var cNewStr = "Cnew"

// implement joint consensus
type configuration struct {
	mutexState // configuration can have C_old state or C_old_new state
	cOldNode   nodeMap
	cNewNode   nodeMap
}

// NewConfig accepts client's given node and return a new configuration
func NewConfig(nodes ...Node) (conf *configuration) {
	conf = &configuration{
		cOldNode: makeNodeMap(nodes...),
	}
	conf.setState(cOld)
	return
}
