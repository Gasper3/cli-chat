package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
)

type clientConn struct {
	conn     net.Conn
	writer   bufio.Writer
	reader   bufio.Reader
	username string
}

func RunServer() {
	clients := make(map[string]clientConn)
	quit := make(chan int)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	port := flag.String("p", "8000", "Application port")
	flag.Parse()

	fmt.Println("Listening on port " + *port)
	listener, err := net.Listen("tcp", ":"+*port)
	HandleError(err, false)

	defer closeListener(&listener)
	go cleanup(sigChan, quit, &listener)

	for {
		conn, err := listener.Accept()
		HandleError(err, false)
		fmt.Println("New connection", conn.RemoteAddr().String())

		client := clientConn{conn: conn, writer: *bufio.NewWriter(conn), reader: *bufio.NewReader(conn)}
		clients[conn.RemoteAddr().String()] = client

		go func(cc clientConn) {
			for {
				msg, err := cc.reader.ReadString('\n')
				if err != nil {
					fmt.Println("Client disconnected:", cc.conn.RemoteAddr().String())
					return
				}
				fmt.Println(fmt.Sprintf("<%s::%s>", cc.conn.RemoteAddr().String(), cc.username), msg)

				if strings.HasPrefix(msg, "/") {
					handleCommand(msg, &cc)
					continue
				}

				for _, cd := range clients {
					if cd.conn != cc.conn {
						cd.writer.WriteString(fmt.Sprintf("%s: %s", strings.Trim(cc.username, "\n"), msg))
						cd.writer.Flush()
					}
				}
			}
		}(client)
	}
}

func handleCommand(cmd string, cc *clientConn) {
	parsedCmd := strings.Split(strings.TrimLeft(cmd, "/"), "::")
	switch parsedCmd[0] {
	case "setusername":
		cc.username = parsedCmd[1]
	}
}

func closeListener(l *net.Listener) {
	fmt.Println("Closing server")
	(*l).Close()
}

func cleanup(c chan os.Signal, quit chan int, listener *net.Listener) {
	for range c {
		fmt.Println("\nExiting...")
		(*listener).Close()
		os.Exit(0)
	}
}
