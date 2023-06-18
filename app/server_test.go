package app

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"testing"
)

type writerMock struct {
	messages []string
}

func (w *writerMock) Write(s []byte) (int, error) {
	w.messages = append(w.messages, string(s))
	return len(s), nil
}

func TestMain(m *testing.M) {
	listener, _ := net.Listen("tcp", "127.0.0.1:8443")
	code := m.Run()
	listener.Close()
	os.Exit(code)
}

func TestHandleCommand(t *testing.T) {
	cc := clientConn{}
	newusername := "NewUsername"
	err := handleCommand(fmt.Sprintf("/setusername::%s\n", newusername), &cc)
	if err != nil {
		t.Fatalf("Expected no error but got %s", err)
	}

	if cc.username != "NewUsername" {
		t.Fatalf("Expected %s got: %s", newusername, cc.username)
	}
}

func FuzzHandleCommand(f *testing.F) {
	f.Add("/setusername::Gasper3")
	f.Fuzz(func(t *testing.T, cmd string) {
		handleCommand(cmd, &clientConn{})
	})
}

func TestBroadcastMessage(t *testing.T) {
	clients := make(clientsMap)
	conn, err := net.Dial("tcp", "127.0.0.1:8443")
	if err != nil {
		t.Error("Connection error", err)
	}

	client := clientConn{conn: conn, username: "Client"}

	var msgs []string
	mock1 := writerMock{messages: msgs}
	mock2 := writerMock{messages: msgs}

	clients["client1"] = &clientConn{writer: *bufio.NewWriter(&mock1)}
	clients["client2"] = &clientConn{writer: *bufio.NewWriter(&mock2)}

	broadcastMessage(client, clients, "New Message")
	if length := len(mock1.messages); length < 1 {
		t.Errorf("Expected 1 message got: %s", fmt.Sprint(length))
	}
	if length := len(mock2.messages); length < 1 {
		t.Errorf("Expected 1 message got: %s", fmt.Sprint(length))
	}
}
