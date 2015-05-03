package lilraft

import (
	"testing"
)

var a []int

func TestNewServer(t *testing.T) {
	NewServer(1, a, nil)
}
