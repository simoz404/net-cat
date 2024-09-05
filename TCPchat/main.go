package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

type Client struct {
	conn net.Conn
	name string
}

var (
	clients    = make(map[*Client]bool)
	clientsMux sync.Mutex
)

func main() {
	port := ":8080"
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("The following error occurred", err)
		return
	}
	fmt.Println("The listener object has been created:", ln)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	client := &Client{conn: conn}
	
	conn.Write([]byte("[ENTER YOUR NAME]: "))
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		client.name = scanner.Text()
	} else {
		return // connection closed or error occurred
	}

	clientsMux.Lock()
	clients[client] = true
	clientsMux.Unlock()

	broadcastMessage(fmt.Sprintf("%s has joined our chat...\n", client.name), client)

	for scanner.Scan() {
		message := scanner.Text()
		broadcastMessage(fmt.Sprintf("%s: %s\n", client.name, message), client)
	}

	clientsMux.Lock()
	delete(clients, client)
	clientsMux.Unlock()

	broadcastMessage(fmt.Sprintf("%s has left the chat...\n", client.name), client)
}

func broadcastMessage(message string, sender *Client) {
	fmt.Print(message) // Print message to server console

	clientsMux.Lock()
	defer clientsMux.Unlock()

	for client := range clients {
		if client != sender { // Don't send the message back to the sender
			_, err := client.conn.Write([]byte(message))
			if err != nil {
				fmt.Printf("Error broadcasting to %s: %v\n", client.name, err)
				client.conn.Close()
				delete(clients, client)
			}
		}
	}
}