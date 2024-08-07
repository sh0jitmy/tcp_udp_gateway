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
    tcpServerAddr = "localhost:12345"
    udpPort       = "12346"
    udpSendAddr   = "localhost:12347"
)

var (
    tcpConn net.Conn
    udpConn *net.UDPConn
    connMu  sync.Mutex
)

func main() {
    // TCP Client
    go startTCPClient()

    // UDP Server
    go startUDPServer()

    // Keep main function running
    for {
        time.Sleep(1 * time.Second)
    }
}

func startTCPClient() {
    for {
        conn, err := net.Dial("tcp", tcpServerAddr)
        if err != nil {
            fmt.Println("Error connecting to TCP server:", err)
            time.Sleep(5 * time.Second) // Retry after a delay
            continue
        }
        connMu.Lock()
        tcpConn = conn
        connMu.Unlock()
        defer conn.Close()
        fmt.Println("Connected to TCP server at", tcpServerAddr)

        reader := bufio.NewReader(conn)
        for {
            message, err := reader.ReadString('\n')
            if err != nil {
                fmt.Println("TCP connection closed:", err)
                connMu.Lock()
                tcpConn = nil
                connMu.Unlock()
                break
            }
            fmt.Println("Received from TCP server:", message)

            // Send message via UDP
            go func(message string) {
                udpAddr, err := net.ResolveUDPAddr("udp", udpSendAddr)
                if err != nil {
                    fmt.Println("Error resolving UDP address:", err)
                    return
                }
                udpConn, err := net.DialUDP("udp", nil, udpAddr)
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
}

func startUDPServer() {
    addr, err := net.ResolveUDPAddr("udp", ":"+udpPort)
    if err != nil {
        fmt.Println("Error resolving UDP address:", err)
        os.Exit(1)
    }

    udpConn, err = net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Error starting UDP server:", err)
        os.Exit(1)
    }
    defer udpConn.Close()
    fmt.Println("UDP server listening on port", udpPort)

    buffer := make([]byte, 1024)
    for {
        n, _, err := udpConn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println("Error reading from UDP connection:", err)
            continue
        }
        message := string(buffer[:n])
        fmt.Println("Received from UDP:", message)

        // Send message to TCP server
        connMu.Lock()
        if tcpConn != nil {
            _, err = tcpConn.Write([]byte(message))
            if err != nil {
                fmt.Println("Error sending to TCP server:", err)
                tcpConn.Close()
                tcpConn = nil
            }
        }
        connMu.Unlock()
    }
}
