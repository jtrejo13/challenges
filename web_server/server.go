package main

import (
	"fmt"
	"net"
	"strings"
	"os"
	"path/filepath"
)

// check panics if an error is not nil
func check(e error) {
    if e != nil {
        panic(e)
    }
}


// loadPageData loads the data for a given page
func loadPageData(path string) []byte {
	
	var filename string
	header := "HTTP/1.1 200 OK\r\n\r\n"

	switch path {
		case "/", "/index.html":
			filename = "index.html"
		default:
			header = "HTTP/1.1 404 Not Found\r\n\r\n"
			filename = "404.html"
	}
	
	pageData, err := os.ReadFile(filepath.Join("pages", filename))
	check(err)

	return append([]byte(header), pageData...)
}


// handleConnection handles incoming HTTP requests
func handleConnection(conn net.Conn) {
	// Close the connection when the function exits
	defer conn.Close()

	// Create a buffer to hold the incoming data
	buffer := make([]byte, 1024)
	// Read the incoming connection into the buffer
	_, err := conn.Read(buffer)
	check(err)

	// Get the first line of the request
	request := string(buffer)
	request = strings.Split(request, "\n")[0]
	// Get the path from the request
	path := strings.Fields(request)[1]

	// Load the data for the requested page
	pageData := loadPageData(path)

	// Send a response back to the person contacting us
	_, err = conn.Write(pageData)
	check(err)
}


func main() {
	// Listen on port 80
	listener, err := net.Listen("tcp", ":80")
	check(err)
	fmt.Println("Server listening on port 80")

	// Close the listener when the application closes
	defer listener.Close()

	// Wait for and accept incoming connections
	for {
		conn, err := listener.Accept()
		check(err)

		// Handle the connection in a new goroutine
		go handleConnection(conn)

	}
}