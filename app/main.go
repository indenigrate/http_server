package main

import (
	"flag"
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
	// Adding Directory Flag
	dir := getFlagString()
	// Listen at 4221
	listenTCP("0.0.0.0:4221", dir)
	fmt.Println("Port 4221 Binded ")

}

func listenTCP(address string, dir string) (net.Listener, error) {
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
		go handleConnection(conn, dir)
	}
	return netListener, nil
}

// handle connection function
func handleConnection(conn net.Conn, dir string) {
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
	} else if targets[1] == "files" {
		fmt.Println("serving /files/{filename}")
		// handle /echo/{filename}

		file, err := os.ReadFile(dir + "/" + targets[2])
		if err != nil {
			fmt.Println("Error opening file:", err)
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
			return
		}
		content := string(file)
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}

	// if target == targetURL {
	// 	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	// } else {
	// 	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	// }
}

// String CLI Input
func getFlagString() string {
	dirAddr := flag.String("directory", "./", "Directory to serve files from")
	flag.Parse()
	return *dirAddr
}
