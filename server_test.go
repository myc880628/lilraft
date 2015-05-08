package lilraft

import (
	"net/http"
	"testing"
)

var testArray []int

func TestNewServer(t *testing.T) {
	logger.Printf("Creating configuration...\n")
	config := NewConfig(
		NewHTTPNode(1, "http://127.0.0.1:8787"),
		NewHTTPNode(2, "http://127.0.0.1:8788"),
		NewHTTPNode(3, "http://127.0.0.1:8789"),
	)
	s1 := NewServer(1, testArray, config)
	s2 := NewServer(2, testArray, config)
	s3 := NewServer(3, testArray, config)
	s1.SetHTTPTransport(http.NewServeMux(), 8787)
	s2.SetHTTPTransport(http.NewServeMux(), 8788)
	s3.SetHTTPTransport(http.NewServeMux(), 8789)
	s1.Start()
	s2.Start()
	s3.Start()
	select {}
}
