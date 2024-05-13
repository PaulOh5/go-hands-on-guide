package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func handleUserAPI(w http.ResponseWriter, r *http.Request) {
	log.Println("I stared processing the request")
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v\n", err)
		http.Error(w, "Error reading body", http.StatusInternalServerError)
		return
	}

	log.Println(string(data))
	fmt.Fprintf(w, "Hello world!\n")
	log.Println("I finished processing the request")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users/", handleUserAPI)

	s := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}
