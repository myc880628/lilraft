package lilraft

import (
	"net/http"
	"testing"
)

var a []int

func TestNewServer(t *testing.T) {
	config := NewConfig(
		NewHTTPNode(1, "http://127.0.0.1:2046"),
		NewHTTPNode(2, "http://127.0.0.1:2047"),
		NewHTTPNode(3, "http://127.0.0.1:2048"),
	)
	s1 := NewServer(1, a, config)
	s2 := NewServer(2, a, config)
	s3 := NewServer(3, a, config)
	s1.SetHTTPTransport(http.NewServeMux(), 2046)
	s2.SetHTTPTransport(http.NewServeMux(), 2047)
	s3.SetHTTPTransport(http.NewServeMux(), 2048)
	s1.Start()
	s2.Start()
	s3.Start()
	select {}
}
