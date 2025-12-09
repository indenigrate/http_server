# Go HTTP/1.1 Server

A lightweight, concurrent HTTP/1.1 server implementation written in Go. This project interacts directly with the TCP layer, handling raw byte streams to parse HTTP requests and generate compliant responses without relying on high-level `net/http` handler abstractions. It is designed to serve static files, echo request data, and handle concurrent client connections via Go routines.

## Capabilities

The server listens on port `4221` and implements the following endpoints and behaviors:

- **`/`**: Returns a standard 200 OK status.
- **`/echo/{string}`**: Returns the path parameter as the response body.
  - Supports [Gzip compression](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding) if the `Accept-Encoding: gzip` header is present.
- **`/user-agent`**: Reads and returns the value of the `User-Agent` header from the client request.
- **`/files/{filename}`**:
  - `GET`: Reads the specified file from the server's configured directory and returns it with `application/octet-stream`.
  - `POST`: Creates a new file (or overwrites an existing one) with the request body data.
- **Persistent Connections**: Supports [HTTP Keep-Alive](https://developer.mozilla.org/en-US/docs/Web/HTTP/Connection_management_in_HTTP_1.x) by processing multiple requests on a single connection loop until a `Connection: close` header is detected.

## Implementation Techniques

The codebase demonstrates several low-level networking and system programming techniques:

- **Raw TCP Socket Management**: Utilizes the Go `net` package to listen on TCP sockets and accept incoming connections, bypassing the standard HTTP multiplexer.
- **Manual Protocol Parsing**: Implements a custom parser using `bufio` to read the [HTTP Request Line](https://developer.mozilla.org/en-US/docs/Web/HTTP/Messages#request_line) and [Headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers) directly from the wire. This includes handling CRLF delimiters and parsing `Content-Length` to delimit body payloads.
- **Concurrency**: Spawns lightweight [Go routines](https://go.dev/tour/concurrency/1) for every incoming connection, ensuring the main listener remains unblocked.
- **Compression Negotiation**: Manually checks `Accept-Encoding` headers and utilizes a `gzip.Writer` buffer to compress response bodies when requested by the client.
- **CLI Flag Parsing**: Uses the `flag` package to accept directory configurations at runtime (e.g., `--directory /tmp`).

## Technologies and Libraries

This project relies on the Go standard library:

- [net](https://pkg.go.dev/net): Provides the portable interface for network I/O, including TCP/IP listeners.
- [bufio](https://pkg.go.dev/bufio): Implements buffered I/O, essential for reading variable-length HTTP lines efficiently.
- [compress/gzip](https://pkg.go.dev/compress/gzip): Implements reading and writing of gzip format compressed files.
- [os](https://pkg.go.dev/os): Provides a platform-independent interface to operating system functionality, used here for file reading/writing.
- [bytes](https://pkg.go.dev/bytes): Implements functions for the manipulation of byte slices, used for buffer management during compression.

## Project Structure

```text
.
└── app/