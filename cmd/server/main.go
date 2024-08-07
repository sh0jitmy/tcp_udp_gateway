package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	tcpPort = "12345"
	udpPort = "12346"
)

var (
	clients   = make(map[net.Conn]bool)
	clientsMu sync.Mutex
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	// TCP Server
	go func() {
		defer wg.Done()
		startTCPServer()
	}()

	// UDP Server
	go func() {
		defer wg.Done()
		startUDPServer()
	}()

	// Wait for both servers to finish
	for {
		time.Sleep(1 * time.Second)
	}
}

func startTCPServer() {
	listener, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("TCP server listening on port", tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			continue
		}
		clientsMu.Lock()
		clients[conn] = true
		clientsMu.Unlock()
		go handleTCPConnection(conn)
	}
}

func handleTCPConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("TCP connection closed:", err)
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
			return
		}
		fmt.Println("Received from TCP:", message)

		// Send message via UDP
		go func(message string) {
			udpConn, err := net.Dial("udp", "localhost:"+udpPort)
			if err != nil {
				fmt.Println("Error connecting to UDP server:", err)
				return
			}
			defer udpConn.Close()
			_, err = udpConn.Write([]byte(message))
			if err != nil {
				fmt.Println("Error sending to UDP server:", err)
			}
		}(message)
	}
}

func startUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", ":"+udpPort)
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("UDP server listening on port", udpPort)

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			continue
		}
		message := string(buffer[:n])
		fmt.Println("Received from UDP:", message)

		// Send message to all TCP clients
		clientsMu.Lock()
		for client := range clients {
			go func(client net.Conn, message string) {
				_, err := client.Write([]byte(message))
				if err != nil {
					fmt.Println("Error sending to TCP client:", err)
					client.Close()
					clientsMu.Lock()
					delete(clients, client)
					clientsMu.Unlock()
				}
			}(client, message)
		}
		clientsMu.Unlock()
	}
}
