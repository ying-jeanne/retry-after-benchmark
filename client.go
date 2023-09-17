package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Send 20 requests per second for 5 minutes
	senario(1, 20, 30*time.Second)
	// Send 200 requests per second for 2 seconds
	senario(2, 2000, 2*time.Second)

	// Send 20 requests per second for 10 minutes
	senario(3, 20, 600*time.Second)
}

func senario(senarioNum int, rate int, duration time.Duration) {
	// Create a channel that will recieve ticks from the ticker
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop() // Stop the ticker when we are done
	counter := 0        // Initialize a counter

	// Create a channel to signal when to stop the ticker
	stopTicker := make(chan bool)

	go func() {
		time.Sleep(duration)
		stopTicker <- true // Signal to stop the ticker after the specified duration
	}()

	for {
		select {
		case <-ticker.C:
			counter++
			go simulateNormalRequest(senarioNum, 0, counter)

		case <-stopTicker:
			// Stop the ticker by breaking out of the loop
			return
		}
	}
}

func simulateNormalRequest(senarioNum int, retryNum int, counter int) {
	responseCode, retryAfter := sendRequest(senarioNum, retryNum, counter)
	if responseCode == http.StatusTooManyRequests {
		// If the response is 429 (rate limit exceeded), add it to the retry channel
		fmt.Printf("Received 429 response for senario %d, normal request %d. Will retry after %s.\n", senarioNum, counter, retryAfter)
		time.Sleep(retryAfter)
		simulateNormalRequest(senarioNum, retryNum+1, counter)
	} else {
		// Decode and process the response (you can replace this with your actual response handling logic)
		processResponse(senarioNum, counter, responseCode)
	}
}

func DoRequest(retryNum int, currentConnect int) *http.Response {
	if currentConnect > 5 {
		return nil
	}
	serverURL := "http://localhost:8080"
	req, err := http.NewRequest("GET", serverURL, nil)

	req.Header.Set("Retry-Count", fmt.Sprintf("%d", retryNum))

	if err != nil {
		log.Fatal(fmt.Println("Error creating request:", err))
	}
	// Send the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		DoRequest(retryNum, currentConnect+1)
	}
	return resp
}

func sendRequest(senarioNum, retryNum int, counter int) (int, time.Duration) {
	resp := DoRequest(retryNum, 0)

	if resp == nil {
		return http.StatusInternalServerError, 0
	}
	defer resp.Body.Close()

	waitTime := 0 * time.Second
	if resp.StatusCode == http.StatusTooManyRequests {
		fmt.Println("Rate limit exceeded. Waiting for Retry-After...")
		retryAfterHeader := resp.Header.Get("Retry-After")
		if retryAfterHeader != "" {
			waitTime, _ = time.ParseDuration(retryAfterHeader + "s")
			// if err != nil {
			// 	log.Fatal(fmt.Println("Error parsing Retry-After header:", err))
			// }
		} else {
			fmt.Println("Maximum retry reached, Retry-After header not found")
			return http.StatusInternalServerError, 0
		}
	}
	return resp.StatusCode, waitTime

	// Simulate a response with a Retry-After header (429 status code for rate limiting)
	// if rand.Intn(10) == 0 {
	// 	retryAfter := time.Duration(rand.Intn(5)+1) * time.Second
	// 	fmt.Printf("Received 429 response on senario %d request No %d. Will retry after %s.\n", senarioNum, counter, retryAfter)
	// 	return http.StatusTooManyRequests, retryAfter
	// }

	// // Simulate a successful response (200 status code) with a random processing time
	// processingTime := time.Duration(rand.Intn(5)) * time.Millisecond
	// time.Sleep(processingTime)
	// return http.StatusOK, 0
}

func processResponse(senarioNum int, counter int, responseCode int) {
	// Simulate response processing based on the status code (you can replace this with your actual response handling logic)
	fmt.Printf("Processing senario %d response %d with status code: %d\n", senarioNum, counter, responseCode)
	// Add your response processing logic here
}
