package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
)


type Request struct {
	ID int
	URL string
}


func makeRequests(id int, requests <-chan Request, results chan<- bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range requests {
		resp, _ := http.Get(req.URL)
		log.Printf("Worker group %d received response: %s", id, resp.Status)
		results <- resp.StatusCode == 200
	}
}

  

func main() {
	
	url := flag.String("u", "", "URL to make a GET request to")
	numRequests := flag.Int("n", 1, "Number of requests to make")
	concurrency := flag.Int("c", 1, "Number of concurrent requests to make")

	flag.Parse()

	requests := make(chan Request, *numRequests)
	results := make(chan bool, *numRequests)


	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go makeRequests(i, requests, results, &wg)
	}

	// Send requests
	for i := 0; i < *numRequests; i++ {
		requests <- Request{ID: i, URL: *url}
	}
	close(requests)


	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Count results
	successCount := 0
    failureCount := 0
	for result := range results {
		if result {
			successCount++
		} else {
			failureCount++
		}
	}

	fmt.Print("All workers finished\n")
	fmt.Printf("Success: %d\n", successCount)
	fmt.Printf("Failure: %d\n", failureCount)
}