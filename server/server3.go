package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

const (
	// Set the rate limit to 5 requests per second with a burst limit of 10 requests.
	rateLimit         = 25
	burst             = 50
	maxRetryAttempts  = 7   // Maximum number of retry attempts.
	initialRetryDelay = 2   // Initial retry delay in seconds.
	maxRetryDelay     = 256 // Maximum retry delay in seconds.
)

var (
	requestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	})
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
			retryCountStr := r.Header.Get("Retry-Count")
			retryCount := 0
			if retryCountStr != "" {
				var err error
				retryCount, err = strconv.Atoi(retryCountStr)
				if err != nil || retryCount < 0 {
					retryCount = 0
					fmt.Println("retryCountStr is not a number")
				}
			}

			if retryCount >= maxRetryAttempts {
				retryCount = maxRetryAttempts - 1
			}

			retryDelay := time.Duration(initialRetryDelay) * time.Second << uint(retryCount)
			if retryDelay > time.Duration(maxRetryDelay)*time.Second {
				retryDelay = time.Duration(maxRetryDelay) * time.Second
			}

			fmt.Println("the retry delay is", retryDelay.Seconds())
			// Set the "Retry-After" header with the time to wait in seconds.
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryDelay.Seconds()))
			w.Header().Set("Retry-Count", fmt.Sprintf("%d", retryCount+1))

			// Return a "Too Many Requests" (HTTP 429) status code.
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
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
