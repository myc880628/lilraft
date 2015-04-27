package lilraft

import "sync"

const (
	follower = iota
	candidate
	leader
)

//-------------------state wraps stateInt with a Mutex--------------------------
type state struct {
	stateInt uint8
	sync.RWMutex
}

func (ss *state) get() uint8 {
	ss.RLock()
	defer ss.RUnlock()
	return ss.stateInt
}

func (ss *state) set(stateInt uint8) {
	ss.Lock()
	defer ss.Unlock()
	ss.stateInt = stateInt
}

type server struct {
	id    uint16
	state *state
}
