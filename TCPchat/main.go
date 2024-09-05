package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	port := ":8080"
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("The following error occured", err)
	} else {
		fmt.Println("The listener object has been created:", ln)
	}
	for {
		con, _ := ln.Accept()
		go connect(con)
	}
}

func connect(con net.Conn) {
	fmt.Println("new user connected")
	con.Write([]byte("[ENTER YOUR NAME]:"))
	scanner := bufio.NewScanner(con)
	index := 0
	for scanner.Scan() {
		message := scanner.Text()
		if index == 0 {
			s := message + " has joined our chat..."
			con.Write([]byte(s))
			fmt.Println(message, " has joined our chat...")
		} else {
		fmt.Println("Received:", message)
			con.Write([]byte(message))
		}
		index++
	}
}
