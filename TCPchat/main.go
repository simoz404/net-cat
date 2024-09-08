package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var color = map[string]string{
	"red":     "\033[31m",
	"green":   "\033[32m",
	"yellow":  "\033[33m",
	"blue":    "\033[34m",
	"magenta": "\033[35m",
	"cyan":    "\033[36m",
	"white":   "\033[37m",
	"orange":  "\033[38;5;208m",
}

var color1 = map[int]string{
	0:     "\033[31m",
	1:   "\033[32m",
	2:  "\033[33m",
	3:    "\033[34m",
	4: "\033[35m",
	5:    "\033[36m",
	6:   "\033[37m",
}

var (
	allMessages   string
	clients       = make(map[net.Conn]string)
	names         = make(map[string]bool)
	clientsMu     sync.RWMutex
	allMessagesMu sync.RWMutex
)

var WelcomeMessage = "Welcome to TCP-Chat!\n         _nnnn_\n        dGGGGMMb\n       @p~qp~~qMb\n       M|@||@) M|\n       @,----.JM|\n      JS^\\__/  qKL\n     dZP        qKRb\n    dZP          qKKb\n   fZP            SMMb\n   HZM            MMMM\n   FqM            MMMM\n __| \".        |\\dS\"qML\n |    `.       | `' \\Zq\n_)      \\.___.,|     .'\n\\____   )MMMMMP|   .'\n     `-'       `--'\n[ENTER YOUR NAME]: "

func main() {
	port := ":8080"
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	if len(os.Args) == 2 {
		port = ":" + os.Args[1]
	}
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()
	fmt.Printf("Server is listening on port %s\n", port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	n:=0
	for _, v := range WelcomeMessage {
			conn.Write([]byte(color1[n]+string(v)+"\033[0m"))
			n++
			if n > len(color1) {
				n=0
			}
	}
	//conn.Write([]byte(WelcomeMessage))
	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error reading client name:", err)
		return
	}
	name = strings.TrimSpace(name)

	clientsMu.Lock()
	if len(names) > 9 {
		clientsMu.Unlock()
		for {
			conn.Write([]byte("The server has reached Maximum 10 connections.\n"))
			return
		}
	} else if names[name] || !validMessage(name) {
		clientsMu.Unlock()
		for {
			conn.Write([]byte("Please choose another name.\n"))
			conn.Write([]byte("[ENTER YOUR NAME]: "))
			name, _ = bufio.NewReader(conn).ReadString('\n')
			name = strings.TrimSpace(name)
			clientsMu.Lock()
			if !names[name] && validMessage(name) {
				names[name] = true
				clientsMu.Unlock()
				break
			}
			clientsMu.Unlock()
		}
	} else {
		names[name] = true
		clientsMu.Unlock()
	}

	clientsMu.Lock()
	clients[conn] = name
	clientsMu.Unlock()

	allMessagesMu.Lock()
	conn.Write([]byte(allMessages))
	allMessagesMu.Unlock()

	sendMessage(fmt.Sprintf(color["yellow"]+"\n%s has joined the chat\n"+"\033[0m", name), conn)
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			fmt.Printf(color["red"]+"Error reading from client %s: %v\n"+"\033[0m", name, err)
			disconnectClient(conn)
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			break
		}
		sendMessage(formatMessage(msg, name), conn)
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
	sendMessage(fmt.Sprintf(color["red"]+"\n%s has left the chat\n"+"\033[0m", name), conn)
}

func sendMessage(msg string, sender net.Conn) {
	saveMessages(msg)
	clientsMu.RLock()
	defer clientsMu.RUnlock()
	msg = "\n" + msg
	for client, name := range clients {
		if client != sender {
			_, err := client.Write([]byte(msg))
			if err != nil {
				fmt.Printf(color["red"]+"Error broadcasting to %s: %v\n"+"\033[0m", name, err)
				client.Close()

			}
		}
		client.Write([]byte(formatMessage("", name)))
	}
}

func formatMessage(message string, name string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s][%s]:%s %s"+"\033[0m", currentTime, name, color["green"], message)
}

func saveMessages(msg string) {
	allMessagesMu.Lock()
	allMessages += msg
	allMessagesMu.Unlock()
}

func validMessage(msg string) bool {
	for _, s := range msg {
		if s > 32 && s < 127 {
			return true
		}
	}
	return false
}

func disconnectClient(conn net.Conn) {
	clientsMu.Lock()
	conn.Close()
	delete(clients, conn)
	clientsMu.Unlock()
}
