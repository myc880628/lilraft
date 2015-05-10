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
	oldleader int32
	newLeader int32
)

var (
	s1      *Server
	s2      *Server
	s3      *Server
	servers = make(map[int32]*Server)
)

var ports = []int{8787, 8788, 8789}

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
	servers[s1.id] = s1
	servers[s2.id] = s2
	servers[s3.id] = s3
}

func TestStartServer(t *testing.T) {
	// s1, s2, s3 is on the same machine, so
	// command just need to be registered once
	s1.RegisterCommand(&testCommand{})

	s1.SetHTTPTransport(http.NewServeMux(), ports[0])
	s2.SetHTTPTransport(http.NewServeMux(), ports[1])
	s3.SetHTTPTransport(http.NewServeMux(), ports[2])

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

func TestLeaderCrash(t *testing.T) {
	for _, server := range servers {
		if server.getState() == leader {
			oldleader = server.id
			server.Stop()
			break
		}
	}

	time.Sleep(600 * time.Millisecond)

	logger.Println("old leader: ", oldleader)

	for _, server := range servers {
		if server.getState() == leader {
			newLeader = server.id
			break
		}
	}

	logger.Println("new leader: ", newLeader)

	if oldleader == newLeader {
		t.Errorf("leader should be different")
	}

}

func TestOldLeaderReconnect(t *testing.T) {
	oldLeaderServer := servers[oldleader]
	oldLeaderServer.Start()
	oldLeaderServer.setState(leader)
	logger.Printf("old leader %d reconnected", oldLeaderServer.id)
	time.Sleep(500 * time.Millisecond)
	if oldLeaderServer.getState() == leader {
		t.Errorf("old leader should stepped down")
	}
}

func TestServerStop(t *testing.T) {
	for _, server := range servers {
		if server.id != newLeader {
			server.Stop()
		}
	}
}
