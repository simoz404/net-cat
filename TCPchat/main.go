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
	"red":    "\033[31m",
	"green":  "\033[32m",
	"yellow": "\033[33m",
}

var (
	allMessages   string
	clients       = make(map[net.Conn]string)
	names         = make(map[string]bool)
	clientsMu     sync.Mutex
	allMessagesMu sync.Mutex
)

var WelcomeMessage = "Welcome to TCP-Chat!\n         _nnnn_\n        dGGGGMMb\n       @p~qp~~qMb\n       M|@||@) M|\n       @,----.JM|\n      JS^\\__/  qKL\n     dZP        qKRb\n    dZP          qKKb\n   fZP            SMMb\n   HZM            MMMM\n   FqM            MMMM\n __| \".        |\\dS\"qML\n |    `.       | `' \\Zq\n_)      \\.___.,|     .'\n\\____   )MMMMMP|   .'\n     `-'       `--'\n[ENTER YOUR NAME]: "

func main() {
	addr := ":8989"

	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	if len(os.Args) == 2 {
		addr = ":" + os.Args[1]
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf(color["red"]+"%s"+"\033[0m"+"\n[USAGE]: ./TCPChat $port\n", err)
		return
	}
	defer ln.Close()
	fmt.Printf("Server is listening on port %s\n", addr)
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

	conn.Write([]byte(WelcomeMessage))

	name, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return
	}
	name = strings.TrimSpace(name)

	if len(clients) > 9 {
		conn.Write([]byte(color["red"] + "The server has reached Maximum 10 connections.\n" + "\033[0m"))
		return
	}

	if names[name] || !validMessage(name) {
		for {
			conn.Write([]byte(color["red"] + "Please choose another name.\n" + "\033[0m"))
			conn.Write([]byte("[ENTER YOUR NAME]: "))
			name, _ = bufio.NewReader(conn).ReadString('\n')
			name = strings.TrimSpace(name)
			if !names[name] && validMessage(name) {
				names[name] = true
				break
			}
		}
	} else {
		names[name] = true
	}

	clientsMu.Lock()
	clients[conn] = name
	clientsMu.Unlock()

	allMessagesMu.Lock()
	conn.Write([]byte(allMessages))
	allMessagesMu.Unlock()

	sendMessage(fmt.Sprintf(color["yellow"]+"%s has joined the chat\n"+"\033[0m", name), conn)
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			break
		}
		sendMessage(formatMessage(msg, name), conn)
	}

	clientsMu.Lock()
	delete(clients, conn)
	delete(names, name)
	clientsMu.Unlock()
	sendMessage(fmt.Sprintf(color["red"]+"%s has left the chat\n"+"\033[0m", name), conn)
}

func sendMessage(msg string, sender net.Conn) {
	saveMessages(msg)
	clientsMu.Lock()
	defer clientsMu.Unlock()
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
