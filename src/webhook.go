package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func StartWebhookReciever() {
	// Start a new HTTP server that listens on port 3000
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalf("Error reading request body: %v", err)
			return
		}

		// Print the request body
		fmt.Println(string(body))

		// Respond with a 200 OK status
		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":3000", nil))
}
