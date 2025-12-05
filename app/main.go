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

	// TODO: Uncomment the code below to pass the first stage
	//
	// Listen at 4221
	l, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Port 4221 Binded ")

	// Accept connection
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Connection Accepted from ", conn.RemoteAddr())

	// Respond with a simple HTTP response
	handleConnection(conn, "/")
	// conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

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
	reqLineParts := strings.Split(reqParts[0], " ")
	// method := reqLineParts[0]

	target := reqLineParts[1]
	if target == targetURL {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	return
}
