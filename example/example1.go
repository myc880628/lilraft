package main

import (
	"net/http"
	"time"

	"github.com/lilwulin/lilraft"
)

func main() {
	var array []int
	config := lilraft.NewConfig(
		lilraft.NewHTTPNode(1, "http://127.0.0.1:8787"),
		lilraft.NewHTTPNode(2, "http://127.0.0.1:8788"),
		lilraft.NewHTTPNode(3, "http://127.0.0.1:8789"),
	)
	s := lilraft.NewServer(1, &array, config, "server/")
	s.SetHTTPTransport(http.DefaultServeMux, 8787)
	s.Start(false)
	time.Sleep(50 * time.Millisecond)
}
