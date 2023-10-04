package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

const (
	// Set the rate limit to 5 requests per second with a burst limit of 10 requests.
	rateLimit = 25
	burst     = 50
)

var (
	requestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	})
	inflightPushRequests atomic.Int64
)

func main() {
	limiter := rate.NewLimiter(rate.Limit(rateLimit), burst)
	prometheus.MustRegister(requestCounter)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is allowed by the rate limiter.
		requestCounter.Inc()
		fmt.Println("Request received")
		if !limiter.Allow() {
			// Calculate the time to wait before retrying based on the rate limit.
			// Generate jitter following a normal distribution
			failedRequestCount := inflightPushRequests.Load() // Cap the base delay to 30 seconds
			fmt.Println("the failedRequestCount is", failedRequestCount)
			if failedRequestCount < 0 {
				failedRequestCount = 0
			}

			delaySeconds := float64(failedRequestCount) * 0.1

			// Add a random jitter between -0.7*delaySeconds and +0.7*delaySeconds.
			if delaySeconds < 5 {
				// set minimum delay to 1 second
				delaySeconds = 5
			} else if delaySeconds > 30 {
				delaySeconds = 30
			}

			jitter := (rand.Float64()*2.0 - 1.0) * delaySeconds
			delaySeconds += jitter
			waitTime := time.Duration(int64(delaySeconds)) * time.Second

			// Set the "Retry-After" header with the time to wait in seconds.
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", waitTime.Seconds()))

			// Return a "Too Many Requests" (HTTP 429) status code.
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			inflightPushRequests.Add(1)
			return
		} else {
			if inflightPushRequests.Load() > 0 {
				inflightPushRequests.Add(-1)
			}
		}

		// Handle your request logic here.
		// In this example, we'll just return "Hello, World!".
		fmt.Fprintln(w, "Hello, World!")
	})

	http.Handle("/metrics", promhttp.Handler())

	// Start the server on port 8080.
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  200 * time.Second,
		WriteTimeout: 200 * time.Second,
	}

	fmt.Println("Server is running on :8080...")
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}

}
