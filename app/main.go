package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"slices"
	"strconv"
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
	// print the complete conn
	// b := make([]byte, 1024)
	// conn.Read(b)
	// fmt.Print(string(b))
	reader := bufio.NewReader(conn)
	for {
		// Build the req string
		var requestBuilder strings.Builder
		var contentLength int

		// --- STEP A: Read Headers Line-by-Line ---
		for {
			// Read until newline
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error reading:", err)
				}
				return
			}

			// Add line to our builder
			requestBuilder.WriteString(line)

			// Clean line to check for end of headers
			trimmedLine := strings.TrimSpace(line)

			// If line is empty, we found the end of the headers (\r\n\r\n)
			if trimmedLine == "" {
				break
			}

			// (Crucial) Check for Content-Length so we know if there is a body
			// We use ToLower so we match "Content-Length" or "content-length"
			if strings.HasPrefix(strings.ToLower(line), "content-length:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					contentLength, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
				}
			}
		}

		// --- STEP B: Read Body (If exists) ---
		// If we don't do this, the next loop iteration will try to read the body as a header!
		if contentLength > 0 {
			bodyBytes := make([]byte, contentLength)
			// ReadFull ensures we get exactly the number of bytes we expect
			_, err := io.ReadFull(reader, bodyBytes)
			if err != nil {
				return
			}
			requestBuilder.Write(bodyBytes)
		}

		// --- STEP C: Create 'req' variable ---
		// Now 'req' contains the full headers + \r\n + body
		req := requestBuilder.String()
		fmt.Print(req)
		reqParts := strings.Split(req, "\r\n")
		// reqParts[0] is the request line
		// reqParts[1] is the first header line
		// reqParts[2] is the second header line
		reqBody := reqParts[len(reqParts)-1]
		reqLineParts := strings.Split(reqParts[0], " ")
		// method := reqLineParts[0]

		targets := strings.Split(reqLineParts[1], "/")

		if targets[1] == "" {
			// handle /
			fmt.Println("serving /")
			// conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
		} else if targets[1] == "echo" {
			fmt.Println("serving /echo/{str}")
			for _, headerLine := range reqParts {
				if strings.HasPrefix(headerLine, "Accept-Encoding:") {
					encodingsAccepted := strings.Split(strings.TrimSpace(strings.TrimPrefix(headerLine, "Accept-Encoding:")), ", ")
					if slices.Contains(encodingsAccepted, "gzip") {
						commpressedString, err := compressString(targets[2])
						if err != nil {
							conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
							return
						}
						conn.Write(fmt.Appendf(nil, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: gzip\r\n\r\n%s", len(commpressedString), commpressedString))
						return
					}
				}
			}
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

				}
			}
		} else if targets[1] == "files" {
			fmt.Println("serving /files/{filename}")
			// handle /files/{filename}
			filename := targets[2]
			fileAddr := dir + "/" + filename
			// handle GET
			if reqLineParts[0] == "GET" {
				file, err := os.ReadFile(fileAddr)
				if err != nil {
					fmt.Println("Error opening file:", err)
					conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

				}
				content := string(file)
				conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), content)))
			} else if reqLineParts[0] == "POST" {
				// handle POST
				err := os.MkdirAll(dir, 0755) // 0755 grants read/write/execute for owner, read/execute for group/others
				if err != nil {
					log.Fatalf("Failed to create directory: %v", err)
				}
				file, err := os.Create(fileAddr)
				if err != nil {
					fmt.Println("Error creating file:", err)
					conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))

				}
				defer file.Close()
				_, err = file.WriteString(strings.TrimRight(reqBody, "\x00"))
				if err != nil {
					fmt.Println("Error writing to file:", err)
					conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))

				}
				conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
			}
		} else {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}

		if slices.Contains(reqParts, "Connection: close") {
			return
		}
	}
	// handleConnection(conn, dir)
}

// String CLI Input
func getFlagString() string {
	dirAddr := flag.String("directory", "./", "Directory to serve files from")
	flag.Parse()
	return *dirAddr
}

func compressString(input string) (string, error) {
	var b bytes.Buffer
	gzWriter := gzip.NewWriter(&b)
	_, err := gzWriter.Write(fmt.Append(nil, input))
	gzWriter.Close()
	if err != nil {
		fmt.Println("Error compressing data:", err)
		return "", err
	}
	return b.String(), nil
}
