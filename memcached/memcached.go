package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"regexp"
	"strconv"
	"time"
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

	setRegex := regexp.MustCompile(`^\bset\b\s+(\w+)\s+(\d+)\s+(-?\d+)\s+(\d+)`)
	replaceRegex := regexp.MustCompile(`^\breplace\b\s+(\w+)\s+(\d+)\s+(-?\d+)\s+(\d+)`)
	getRegex := regexp.MustCompile(`^\bget\b\s+(\w+)`)

	if setRegex.MatchString(request) {
		return "set", setRegex.FindStringSubmatch(request), true
	} else if getRegex.MatchString(request) {
		return "get", getRegex.FindStringSubmatch(request), true
	} else if replaceRegex.MatchString(request) {
		return "replace", replaceRegex.FindStringSubmatch(request), true
	}

	log.Printf("Invalid request: %s\n", request)
	return "", nil, false
}

// set stores a key-value pair in the cache
// The key is the key to store the value under
// The value is the value to store
// The flags are metadata about the value
// The exptime is the time in seconds until the value expires
// The byteCount is the number of bytes in the value
// The allowReplacement flag determines if the key-value pair can be replaced
// Returns true if the key-value pair was stored successfully
func set(key string, value string, flags string, exptime string, byteCount string, allowReplacement bool) (bool) {

	// Check if the key already exists and if we are allowed to replace it
	if _, ok := globalCache[key]; ok && !allowReplacement {
		log.Printf("Key already exists: %s\n", key)
		return false
	}

	if exptime == "0" {
		exptime = fmt.Sprint(math.MaxInt64)
		log.Printf("Setting key: %s, with value: %s and no exptime\n", key, value)
	} else {
		now := time.Now().Unix()
		exptimeInt, _ := strconv.ParseInt(exptime, 10, 64)
		exptime = fmt.Sprint(now + exptimeInt)
		log.Printf("Setting key: %s, with value: %s and exptime: %s\n", key, value, exptime)
	}

	
	globalCache[key] = []string{value, flags, exptime, byteCount}

	return true
}

func get(key string) []string {
	if value, ok := globalCache[key]; !ok {
		return nil
	} else {

		// Check if the key has expired
		exptime, _ := strconv.Atoi(value[2])
		if int64(exptime) < time.Now().Unix() {
			log.Printf("Key has expired: %s\n", key)
			delete(globalCache, key)
			return nil
		}

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
				case "set", "replace":
					log.Printf("Received %s request for key: %s\n", command, key)

					flags := args[2]
					exptime := args[3]
					byteCount := args[4]

					allowReplacement := command == "replace"

					// Read the next line of the request
					scanner.Scan()
					value := scanner.Text()

					if command == "replace" {
						if _, ok := globalCache[key]; !ok {
							conn.Write([]byte("NOT_STORED\r\n"))
							continue
						}
					}

					if ok := set(key, value, flags, exptime, byteCount, allowReplacement); !ok {
						conn.Write([]byte("NOT_STORED\r\n"))
						continue
					}

					log.Printf("Stored value: %s\n", value)

					// Send a response to the client
					conn.Write([]byte("STORED\r\n"))
				case "get":
					log.Printf("Received get request for key: %s\n", key)

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

