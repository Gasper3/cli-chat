package main

import (
	"bufio"
	"chat-app/utils"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

type clientConn struct {
	conn      net.Conn
	writer    bufio.Writer
	reader    bufio.Reader
	username  string
	sendColor lipgloss.Style
}

func main() {
    RunServer()
}

func (cc *clientConn) formatedMessage(s string) string {
	return fmt.Sprintf("%s %s", cc.sendColor.Render(cc.username+":"), s)
}

func (cc *clientConn) sendToClient(s string) error {
	s = prepareMessage(s)
	cc.writer.WriteString(s)
	err := cc.writer.Flush()
	return err
}

func RunServer() {
	var clients sync.Map

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	port := flag.String("p", "8000", "Application port")
	flag.Parse()

	log.Println("Listening on port " + *port)
	listener, err := net.Listen("tcp", ":"+*port)
	utils.FatalOnError(err)

	go cleanup(sigChan, &listener)

	colorNr := 1
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}

		log.Println("New connection", conn.RemoteAddr().String())
		client := clientConn{
			conn:      conn,
			writer:    *bufio.NewWriter(conn),
			reader:    *bufio.NewReader(conn),
			sendColor: lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprint(colorNr))),
		}
		colorNr++

		clients.Store(conn.RemoteAddr().String(), &client)

		go handleClientConnection(&client, &clients)
	}
}

func handleClientConnection(cc *clientConn, clients *sync.Map) {
	for {
		msg, err := cc.reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				disconectClient(cc, clients, fmt.Sprintf("%s disconnected", cc.username))
				return
			}
			log.Println(err)
		}

		if strings.HasPrefix(msg, "/") {
			err := handleCommand(msg, cc)
			if e, ok := err.(*utils.ClientLeftError); ok && e != nil {
				disconectClient(cc, clients, fmt.Sprint(e))
				return
			}
			if err != nil {
				cc.sendToClient(fmt.Sprint(err))
			}
			continue
		}

		broadcastMessage(*cc, clients, cc.formatedMessage(msg))
	}
}

func disconectClient(cc *clientConn, clients *sync.Map, msg string) {
	log.Println("Client disconnected:", cc.conn.RemoteAddr().String())
	clients.Delete(cc.conn.RemoteAddr().String())
	broadcastMessage(*cc, clients, msg)
}

func broadcastMessage(cc clientConn, clients *sync.Map, msg string) {
	msg = prepareMessage(msg)
	clients.Range(func(key any, value any) bool {
		if c, ok := value.(*clientConn); ok && c.conn != cc.conn {
			if err := c.sendToClient(msg); err != nil {
				log.Println("Broadcast flush error:", err)
			}
		}
		return true
	})
}

func prepareMessage(msg string) string {
	if !strings.HasSuffix(msg, "\n") {
		msg = msg + "\n"
	}
	return msg
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
			return &utils.NotEnoughArgs{CmdName: cmdName}
		}
		if cc.username != "" {
			return errors.New("username already set")
		}
		cc.username = args[0]
		return nil
	case "/leave":
		return &utils.ClientLeftError{Username: cc.username}
	}
	return nil
}

func cleanup(c chan os.Signal, listener *net.Listener) {
	<-c
	log.Println("\nExiting...")
	(*listener).Close()
	os.Exit(0)
}
