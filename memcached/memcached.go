package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"regexp"
)

// check panics if an error is not nil
func check(e error) {
    if e != nil {
        log.Fatal(e)
    }
}

var globalCache = make(map[string][]string)


// validateRequest checks if a request is valid
// A valid request is either a set or get request
// A set request is of the form: set <key> <flags> <exptime> <bytes>
// A get request is of the form: get <key>
func validateRequest(request string) (command string, args []string, valid bool) {
	log.Printf("Validating request: %s\n", request)

	setRegex := regexp.MustCompile(`^\bset\b\s+(\w+)\s+(\d+)\s+(\d+)\s+(\d+)`)
	getRegex := regexp.MustCompile(`^\bget\b\s+(\w+)`)

	if setRegex.MatchString(request) {
		return "set", setRegex.FindStringSubmatch(request), true
	} else if getRegex.MatchString(request) {
		return "get", getRegex.FindStringSubmatch(request), true
	}

	return "", nil, false
}

func set(key string, value string, flags string, exptime string, byteCount string) {
		globalCache[key] = []string{value, flags, exptime, byteCount}
}

func get(key string) []string {
	if value, ok := globalCache[key]; !ok {
		return nil
	} else {
		log.Printf("Retrieved value: %s\n", value)
		return value
	}
}

// handleRequest handles incoming requests
func handleRequest(conn net.Conn) {
	for {
		// Create a new Scanner for the incoming connection
		scanner := bufio.NewScanner(conn)
		
		// Read the first line of the request
		scanner.Scan()
		header := scanner.Text()

		// Check if the request is valid
		command, args, isValidRequest := validateRequest(header)

		if isValidRequest {

			key := args[1]

			switch command {
				case "set":

					flags := args[2]
					exptime := args[3]
					byteCount := args[4]

					// Read the next line of the request
					scanner.Scan()
					value := scanner.Text()

					set(key, value, flags, exptime, byteCount)

					log.Printf("Stored value: %s\n", value)

					// Send a response to the client
					conn.Write([]byte("STORED\r\n"))
				case "get":
					// Get the value from the cache
					result := get(key)

					if result == nil {
						conn.Write([]byte("END\r\n"))
						continue
					}

					value := result[0]
					flags := result[1]
					byteCount := result[3]
					conn.Write([]byte(fmt.Sprintf("VALUE %s %s %s\r\n%s\r\n", key, flags, byteCount, value)))
				default:
					log.Printf("Invalid command: %s\n", command)
			}
		}

	}
}

func main() {

	port := flag.String("port", "11211", "Port to run the server on")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	check(err)
	log.Printf("Server listening on port %s...", *port)

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

