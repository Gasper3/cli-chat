package app

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type clientConn struct {
	conn      net.Conn
	writer    bufio.Writer
	reader    bufio.Reader
	username  string
	sendColor lipgloss.Style
}

type clientLeftError struct {
	message  string
	username string
}

func (e *clientLeftError) Error() string {
	return fmt.Sprintf("%s left the chat", e.username)
}

type notEnoughArgs struct {
	cmdName string
}

func (e *notEnoughArgs) Error() string {
	return fmt.Sprintf("Command %s requires argument after '::'", e.cmdName)
}

type clientsMap map[string]*clientConn

func RunServer() {
	clients := make(clientsMap)
	quit := make(chan int)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	port := flag.String("p", "8000", "Application port")
	flag.Parse()

	fmt.Println("Listening on port " + *port)
	listener, err := net.Listen("tcp", ":"+*port)
	HandleError(err)

	defer closeListener(&listener)
	go cleanup(sigChan, quit, &listener)

	colorNr := 1
	for {
		conn, err := listener.Accept()
		HandleError(err)
		fmt.Println("New connection", conn.RemoteAddr().String())

		client := clientConn{
			conn:      conn,
			writer:    *bufio.NewWriter(conn),
			reader:    *bufio.NewReader(conn),
			sendColor: lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprint(colorNr))),
		}
		colorNr++
		clients[conn.RemoteAddr().String()] = &client

		go handleClientConnection(&client, &clients)
	}
}

func handleClientConnection(cc *clientConn, clients *clientsMap) {
	for {
		msg, err := cc.reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", cc.conn.RemoteAddr().String())
			delete(*clients, cc.conn.RemoteAddr().String())
			return
		}
		log.Println(fmt.Sprintf("<%s::%s>", cc.conn.RemoteAddr().String(), cc.username), msg)

		if strings.HasPrefix(msg, "/") {
			fmt.Println("handleClient", &cc)
			err := handleCommand(msg, cc)
			if e, ok := err.(*clientLeftError); ok {
				delete(*clients, cc.conn.RemoteAddr().String())
				broadcastMessage(*cc, *clients, e.message)
				return
			}
			continue
		}

		broadcastMessage(*cc, *clients, msg)
	}
}

func broadcastMessage(cc clientConn, clients clientsMap, msg string) {
	for _, c := range clients {
		if c.conn != cc.conn {
			c.writer.WriteString(fmt.Sprintf("%s %s", cc.sendColor.Render(cc.username, ":"), msg))
			err := c.writer.Flush()
			HandleError(err)
		}
	}
}

func handleCommand(cmd string, cc *clientConn) error {
	cmdSlice := strings.Split(cmd, "::")
	var parsedCmd []string
	for _, s := range cmdSlice {
		parsedCmd = append(parsedCmd, strings.Trim(s, "\n"))
	}
	cmdName := parsedCmd[0]
	args := parsedCmd[1:]

	switch cmdName {
	case "/setusername":
		if len(parsedCmd) == 1 {
			return &notEnoughArgs{cmdName: cmdName}
		}
		cc.username = args[0]
		return nil
	case "/leave":
		return &clientLeftError{username: cc.username}
	}
	return nil
}

func closeListener(l *net.Listener) {
	fmt.Println("Closing server")
	(*l).Close()
}

func cleanup(c chan os.Signal, quit chan int, listener *net.Listener) {
	<-c
	fmt.Println("\nExiting...")
	(*listener).Close()
	os.Exit(0)
}
