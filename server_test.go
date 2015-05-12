package lilraft

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var ( // context for servers
	testArray1 []int
	testArray2 []int
	testArray3 []int
	testArray4 []int
	testArray5 []int
)

var (
	oldleader int32
	newLeader int32
)

var (
	s1      *Server
	s2      *Server
	s3      *Server
	s4      *Server
	s5      *Server
	servers = make(map[int32]*Server)
)

var ports = []int{8787, 8788, 8789, 8790, 8791}

func init() {
	testArray1 = make([]int, 0)
	testArray2 = make([]int, 0)
	testArray3 = make([]int, 0)
	testArray4 = make([]int, 0)
	testArray5 = make([]int, 0)
}

func TestNewServer(t *testing.T) {
	logger.Printf("Creating configuration...\n")
	config := NewConfig(
		NewHTTPNode(1, "http://127.0.0.1:8787"),
		NewHTTPNode(2, "http://127.0.0.1:8788"),
		NewHTTPNode(3, "http://127.0.0.1:8789"),
	)
	s1 = NewServer(1, &testArray1, config, "s1/", false)
	time.Sleep(50 * time.Millisecond)
	s2 = NewServer(2, &testArray2, config, "s2/", false)
	time.Sleep(50 * time.Millisecond)
	s3 = NewServer(3, &testArray3, config, "s3/", false)
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
	s3.Exec(newTestCommand(2046))
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

func TestChangeConfig(t *testing.T) {
	config := NewConfig(
		NewHTTPNode(1, "http://127.0.0.1:8787"),
		NewHTTPNode(2, "http://127.0.0.1:8788"),
		NewHTTPNode(3, "http://127.0.0.1:8789"),
		NewHTTPNode(4, "http://127.0.0.1:8790"),
		NewHTTPNode(5, "http://127.0.0.1:8791"),
	)

	s4 = NewServer(4, &testArray4, config, "s4/", false)
	time.Sleep(50 * time.Millisecond)
	s5 = NewServer(5, &testArray5, config, "s5/", false)

	servers[4] = s4
	servers[5] = s5

	s4.SetHTTPTransport(http.NewServeMux(), ports[3])
	s5.SetHTTPTransport(http.NewServeMux(), ports[4])

	s4.Start()
	s5.Start()

	err := s3.SetConfig(
		NewHTTPNode(1, "http://127.0.0.1:8787"),
		NewHTTPNode(2, "http://127.0.0.1:8788"),
		NewHTTPNode(3, "http://127.0.0.1:8789"),
		NewHTTPNode(4, "http://127.0.0.1:8790"),
		NewHTTPNode(5, "http://127.0.0.1:8791"),
	)
	if err != nil {
		logger.Println("s3 set config error: ", err.Error())
	}
	time.Sleep(500 * time.Millisecond)
	if len(s3.theOtherNodes()) != 4 {
		t.Errorf("node count doesn't match")
	}
	if s3.config.getState() != cOld {
		t.Errorf("config state is wrong")
	}
}

func TestExecCommandAgain(t *testing.T) {
	s4.Exec(newTestCommand(1997))
	time.Sleep(100 * time.Millisecond)
	if testArray1[1] != 1997 {
		t.Errorf("value should be in s1 context")
	}
	if testArray2[1] != 1997 {
		t.Errorf("value should be in s2 context")
	}
	if testArray3[1] != 1997 {
		t.Errorf("value should be in s3 context")
	}
	if testArray4[1] != 1997 {
		t.Errorf("value should be in s2 context")
	}
	if testArray5[1] != 1997 {
		t.Errorf("value should be in s3 context")
	}
}

func TestServerRecover(t *testing.T) {
	testArray := make([][]int, 6)
	for _, server := range servers {
		server.Stop()
	}
	config := NewConfig(
		NewHTTPNode(1, "http://127.0.0.1:8787"),
		NewHTTPNode(2, "http://127.0.0.1:8788"),
		NewHTTPNode(3, "http://127.0.0.1:8789"),
		NewHTTPNode(4, "http://127.0.0.1:8790"),
		NewHTTPNode(5, "http://127.0.0.1:8791"),
	)

	for id := range servers {
		testArray[id] = make([]int, 0)
	}

	for id, server := range servers {
		server = NewServer(id, &testArray[id], config, fmt.Sprintf("s%d/", id), true)
		server.Start()
	}

	time.Sleep(80 * time.Millisecond)

	for id := range servers {
		if testArray[id][1] != 1997 {
			t.Errorf("value should be in s1 context")
		}
		if err := os.RemoveAll(fmt.Sprintf("s%d", id)); err != nil {
			logger.Println(err.Error())
		}
	}

}

func TestServerFinalStop(t *testing.T) {
	for _, server := range servers {
		server.Stop()
	}
}
