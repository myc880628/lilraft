package lilraft

const (
	c_old = iota
	c_old_new
)

// implement joint consensus
type configuration struct {
	mutexState // configuration can have C_old state or C_old_new state
	c_OldNode  nodeMap
	c_NewNode  nodeMap
}

// TODO: fill this
func NewConfig(nodes ...Node) (conf *configuration) {
	conf = &configuration{
		c_OldNode: makeNodeMap(nodes...),
	}
	conf.setState(c_old)
	return
}

// func (c *configuration) allNodes() nodeMap {
// 	allNodeMap := make(nodeMap)
// 	for id, node := range c.c_NewNode {
// 		allNodeMap[id] = node
// 	}

// 	for id, node := range c.c_OldNode {
// 		allNodeMap[id] = node
// 	}
// 	return allNodeMap
// }

// TODO: fill the pass function
// func (c *configuration) pass() bool {

// 	return true
// }

// func (c *configuration) setNode(nodes ..node) {

// }
