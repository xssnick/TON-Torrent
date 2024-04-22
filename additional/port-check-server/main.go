package main

import (
	"log"
	"net"
	"strings"
)

func main() {
	listen, err := net.Listen("tcp", "0.0.0.0:9099")
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()

	log.Println("started on tcp 9099")
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		go handle(conn)
	}
}

func handle(conn net.Conn) {
	log.Println("conn from", conn.RemoteAddr().String())
	defer conn.Close()

	buffer := make([]byte, 8)
	_, err := conn.Read(buffer)
	if err != nil {
		return
	}

	if string(buffer[:2]) == "ME" {
		log.Println("check request from", conn.RemoteAddr().String())

		conn.Write([]byte("OK"))

		addr := strings.SplitN(conn.RemoteAddr().String(), ":", 2)
		nConn, err := net.Dial("tcp", addr[0]+":18889")
		if err == nil {
			nConn.Write([]byte(addr[0]))
			nConn.Close()
		}
	}
}
