package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
)

func RunServer() {
	clients := make(map[string]net.Conn)
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
		clients[conn.RemoteAddr().String()] = conn

		go func(conn net.Conn) {
			for {
				msg, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Println("Client disconnected:", conn.RemoteAddr().String())
					return
				}

				fmt.Println("Client msg:", msg)
				for _, c := range clients {
					if c != conn {
						c.Write([]byte(msg))
					}
				}
			}
		}(conn)
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
