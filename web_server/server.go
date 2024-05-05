package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// function to check if a string is a valid HTTP GET request
func isValidGETRequest(request string) bool {
	// Define a regex pattern for a simple HTTP GET request
	// This pattern checks for the GET method, a path (which is simplified here),
	// the HTTP version, and optionally checks for a "Host" header
	// Note: This is a simplified version and not robust for all HTTP scenarios
	pattern := `(?m)^GET\s+\/[^\s]*\s+HTTP\/1\.[01]$(?:\r?\nHost:\s+[^\s]+)?`

	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	// Check if the request matches the pattern
	return re.MatchString(request)
}


// handleRequest handles incoming HTTP requests
func handleRequest(conn net.Conn) {
	
	defer conn.Close()

	// Create a new Scanner for the incoming connection
	scanner := bufio.NewScanner(conn)

	// Read the first line of the request
	scanner.Scan()
	request := scanner.Text()

	// Check if the request is valid GET request
	if !isValidGETRequest(request) {
		log.Printf("Invalid request: %s\n", request)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	log.Printf("Received request: %s\n", request)

	// Get the path from the request
	path := strings.Fields(request)[1]

	// Load the data for the requested page
	pageData := loadPageData(path)

	// Send a response back to the person contacting us
	conn.Write(pageData)

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
		go handleRequest(conn)

	}
}