package lilraft

import (
	"testing"
)

var a []int

func TestNewServer(t *testing.T) {
	_ = NewServer(1, a, NewConfig(
		NewHTTPNode(1, "http://127.0.0.1:2046"),
		NewHTTPNode(2, "http://127.0.0.1:2047"),
		NewHTTPNode(3, "http://127.0.0.1:2048"),
	))
}
