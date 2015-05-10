package lilraft

import (
	"net/http"
	"testing"
	"time"
)

var ( // context for servers
	testArray1 []int
	testArray2 []int
	testArray3 []int
)

var (
	s1 *Server
	s2 *Server
	s3 *Server
)

func init() {
	testArray1 = make([]int, 0)
	testArray2 = make([]int, 0)
	testArray3 = make([]int, 0)
}

func TestNewServer(t *testing.T) {
	logger.Printf("Creating configuration...\n")
	config := NewConfig(
		NewHTTPNode(1, "http://127.0.0.1:8787"),
		NewHTTPNode(2, "http://127.0.0.1:8788"),
		NewHTTPNode(3, "http://127.0.0.1:8789"),
	)
	s1 = NewServer(1, &testArray1, config)
	s2 = NewServer(2, &testArray2, config)
	s3 = NewServer(3, &testArray3, config)
}

func TestStartServer(t *testing.T) {
	// s1, s2, s3 is on the same machine, so
	// command just need to be registered once
	s1.RegisterCommand(&testCommand{})

	s1.SetHTTPTransport(http.NewServeMux(), 8787)
	s2.SetHTTPTransport(http.NewServeMux(), 8788)
	s3.SetHTTPTransport(http.NewServeMux(), 8789)

	s1.Start()
	s2.Start()
	s3.Start()

	time.Sleep(1 * time.Second)
}

func TestExecCommand(t *testing.T) {
	s3.Exec(newTestCommand(1, 2046))
	time.Sleep(500 * time.Millisecond)

	if testArray1[0] != 2046 {
		t.Errorf("value should be in s1 context")
	}
	if testArray2[0] != 2046 {
		t.Errorf("value should be in s2 context")
	}
	if testArray3[0] != 2046 {
		t.Errorf("value should be in s3 context")
	}

	logger.Println(testArray1)
	logger.Println(testArray2)
	logger.Println(testArray3)
}

func TestServerStop(t *testing.T) {
	var oldleader int32
	if s1.getState() == leader {
		oldleader = s1.id
		s1.Stop()
	} else if s2.getState() == leader {
		oldleader = s2.id
		s2.Stop()
	} else if s3.getState() == leader {
		oldleader = s3.id
		s3.Stop()
	}
	time.Sleep(600 * time.Millisecond)

	logger.Println("old leader: ", oldleader)
	var newLeader int32
	if s1.getState() == leader {
		newLeader = s1.id
		s1.Stop()
	} else if s2.getState() == leader {
		newLeader = s2.id
		s2.Stop()
	} else if s3.getState() == leader {
		newLeader = s3.id
		s3.Stop()
	}
	logger.Println("new leader: ", newLeader)

	if oldleader == newLeader {
		t.Errorf("leader should be different")
	}
}
