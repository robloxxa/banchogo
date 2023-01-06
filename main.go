package main

import (
	"bufio"
	"fmt"
	"net"
	"net/textproto"
)

func connect() net.Conn {
	return conn
}

func disconnect(conn net.Conn) {
	conn.Close()
}

func logon(conn net.Conn) {
	fmt.Fprintf(conn, "USER robloxxa\r\n")
}

func sendData(conn net.Conn, message string) {
}

func main() {
	conn := connect()
	defer disconnect(conn)
	sendData(conn, "USER robloxxa")
	sendData(conn, "PASS b6529d77")
	tp := textproto.NewReader(bufio.NewReader(conn))
	for {
		status, err := tp.ReadLine()
		if err != nil {
			panic(err)
		}
		fmt.Println(status)
	}

}
