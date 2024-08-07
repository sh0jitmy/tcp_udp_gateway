# 概要
tcp udp 変換プログラム


## プロンプト

TCPクライアントとして動作するプログラムに置き換えるには、以下のように修正します。TCPクライアントがサーバに接続し、データの送受信を行います。その後、UDPで受信したデータをTCPサーバに送信します。

以下に示すのは、TCPクライアントがTCPサーバに接続し、UDPで受信したデータをTCPサーバに送信するプログラムの例です。

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "sync"
)

const (
    tcpServerAddr = "localhost:12345"
    udpPort       = "12346"
)

var (
    tcpConn net.Conn
    udpConn *net.UDPConn
    connMu  sync.Mutex
)

func main() {
    var wg sync.WaitGroup
    wg.Add(2)

    // TCP Client
    go func() {
        defer wg.Done()
        startTCPClient()
    }()

    // UDP Server
    go func() {
        defer wg.Done()
        startUDPServer()
    }()

    // Wait for both to finish
    wg.Wait()
}

func startTCPClient() {
    conn, err := net.Dial("tcp", tcpServerAddr)
    if err != nil {
        fmt.Println("Error connecting to TCP server:", err)
        os.Exit(1)
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
            return
        }
        fmt.Println("Received from TCP server:", message)
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
```

### 説明

1. `startTCPClient`関数では、TCPクライアントとしてサーバに接続し、接続が維持されている間、サーバからメッセージを受信し続けます。
2. `startUDPServer`関数では、UDPサーバとして機能し、メッセージを受信します。受信したメッセージはTCPサーバに送信されます。
3. `connMu`というミューテックスを使用して、`tcpConn`のアクセスを同期し、安全にアクセスできるようにします。

このプログラムでは、TCPクライアントがTCPサーバに接続し、UDPで受信したメッセージをTCPサーバに送信する動作を実現しています。


了解しました。`startTCPClient`でサーバから受信したメッセージをUDPで送信するように追加します。以下の修正版をご確認ください：

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "sync"
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
    var wg sync.WaitGroup
    wg.Add(2)

    // TCP Client
    go func() {
        defer wg.Done()
        startTCPClient()
    }()

    // UDP Server
    go func() {
        defer wg.Done()
        startUDPServer()
    }()

    // Wait for both to finish
    wg.Wait()
}

func startTCPClient() {
    conn, err := net.Dial("tcp", tcpServerAddr)
    if err != nil {
        fmt.Println("Error connecting to TCP server:", err)
        os.Exit(1)
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
            return
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
```

### 変更点

1. `startTCPClient`関数でTCPサーバからメッセージを受信した際に、UDP経由で別のアドレス（`udpSendAddr`）にメッセージを送信するように修正しました。
2. UDP送信のためのアドレスを`udpSendAddr`として追加しました。

これで、TCPサーバから受信したメッセージがUDPで送信されるようになります。


了解しました。`startTCPClient`関数内で`net.Dial`が失敗した場合に、プログラムが終了せず、接続が確立するまで継続的に再試行するように修正します。

以下の修正版コードをご確認ください：

```go
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
    var wg sync.WaitGroup
    wg.Add(2)

    // TCP Client
    go func() {
        defer wg.Done()
        startTCPClient()
    }()

    // UDP Server
    go func() {
        defer wg.Done()
        startUDPServer()
    }()

    // Wait for both to finish
    wg.Wait()
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
```

### 変更点

1. `startTCPClient`関数内で`net.Dial`が失敗した場合、`time.Sleep(5 * time.Second)`を使って5秒待機し、その後再試行します。
2. TCP接続が確立されるまでこの処理をループ内で継続します。

これにより、TCPサーバへの接続が確立されるまで再試行し続けることができます。




了解しました。`sync.WaitGroup`を使用せず、`main`関数で1秒ごとにスリープする`for`ループを追加します。これにより、メインゴルーチンが終了せずに、TCPクライアントとUDPサーバが動作し続けるようになります。

以下に修正したコードを示します：

```go
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
```

### 変更点

1. `sync.WaitGroup`を使用せず、`main`関数で`for`ループを使って1秒ごとにスリープします。
2. `startTCPClient`と`startUDPServer`をそれぞれのゴルーチンで実行します。

これで、メインゴルーチンが終了せず、TCPクライアントとUDPサーバが動作し続けるようになります。
