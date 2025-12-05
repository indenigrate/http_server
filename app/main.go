package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	// Listen at 4221
	listenTCP("0.0.0.0:4221")
	// l, err := net.Listen("tcp", "0.0.0.0:4221")

	// if err != nil {
	// 	fmt.Println("Failed to bind to port 4221")
	// 	os.Exit(1)
	// }
	fmt.Println("Port 4221 Binded ")

	// handle multiple connections

	// Accept connection
	// conn, err := l.Accept()
	// if err != nil {
	// 	fmt.Println("Error accepting connection: ", err.Error())
	// 	os.Exit(1)
	// }
	// fmt.Println("Connection Accepted from ", conn.RemoteAddr())

	// Respond with a simple HTTP response
	// go handleConnection(conn, "/")
	// conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

}

func listenTCP(address string) (net.Listener, error) {
	netListener, err := net.Listen("tcp", address)
	defer netListener.Close()
	if err != nil {
		return nil, err
	}
	for true {
		conn, err := netListener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, "/")
	}
	return netListener, nil
}

// handle connection function
func handleConnection(conn net.Conn, targetURL string) {
	defer conn.Close()
	b := make([]byte, 1024)

	_, err := conn.Read(b)
	if err != nil {
		fmt.Println("Error reading response: ", err.Error())
		return
	}

	req := string(b)
	reqParts := strings.Split(req, "\r\n")
	// reqParts[0] is the request line
	// reqParts[1] is the first header line
	// reqParts[2] is the second header line

	reqLineParts := strings.Split(reqParts[0], " ")
	// method := reqLineParts[0]

	targets := strings.Split(reqLineParts[1], "/")

	if targets[1] == "" {
		// handle /
		fmt.Println("serving /")
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if targets[1] == "echo" {
		fmt.Println("serving /echo/{str}")
		// handle /echo/{str}
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(targets[2]), targets[2])))
	} else if targets[1] == "user-agent" {
		fmt.Println("serving /user-agent")
		// handle /user-agent
		// respond with the User-Agent header value
		// loop req parts to find User-Agent header
		for _, headerLine := range reqParts {
			if strings.HasPrefix(headerLine, "User-Agent:") {
				// ua is the user agent value
				ua := strings.TrimSpace(strings.TrimPrefix(headerLine, "User-Agent:"))
				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(ua), ua)))
				return
			}
		}
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

	// if target == targetURL {
	// 	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	// } else {
	// 	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	// }
}
