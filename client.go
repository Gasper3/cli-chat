package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

func RunClient() {
	addr := flag.String("h", "127.0.0.1", "Host address")
	port := flag.String("p", "8000", "Port")
	flag.Parse()

	c, err := net.Dial("tcp", *addr+":"+*port)
	HandleError(err)

	for {
		go func(conn net.Conn) {
			for {
				buf := make([]byte, 1024)
				_, err := conn.Read(buf)
				if err != nil {
					fmt.Println("Server closed...")
					return
				}
				// serverMsg, _ := bufio.NewReader(conn).ReadString('\n')
				fmt.Println("-->", string(buf))
			}
		}(c)

		reader := bufio.NewReader(os.Stdin)

		fmt.Print(">> ")
		msg, _ := reader.ReadString('\n')
		fmt.Fprintf(c, msg+"\n")
	}
}

func ConnectToServer() (net.Conn, error) {
	addr := flag.String("h", "127.0.0.1", "Host address")
	port := flag.String("p", "8000", "Port")
	flag.Parse()

	return net.Dial("tcp", *addr+":"+*port)
}
