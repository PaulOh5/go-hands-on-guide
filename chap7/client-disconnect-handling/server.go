package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	log.Println("ping: Got a request")
	time.Sleep(10 * time.Second)
	fmt.Fprintf(w, "pong")
}

func doSomeWork() {
	time.Sleep(15 * time.Second)
}

func handleUserAPI(w http.ResponseWriter, r *http.Request) {
	done := make(chan bool)

	log.Println("I started processing the request")

	go func() {
		doSomeWork()
		done <- true
	}()

	select {
	case <-done:
		log.Println("doSomeWork done: Continuing request processing")
	case <-r.Context().Done():
		log.Printf("Aborting request processing: %v\n", r.Context().Err())
		return
	}

	log.Println("I finished processing the request")
}

func setupHandlers(mux *http.ServeMux) {
	timeoutDuration := 30 * time.Second
	userHandler := http.HandlerFunc(handleUserAPI)
	hTimeout := http.TimeoutHandler(
		userHandler,
		timeoutDuration,
		"I ran out of time\n",
	)
	mux.Handle("/api/users/", hTimeout)
	mux.HandleFunc("/ping", handlePing)
}

func main() {
	mux := http.NewServeMux()
	setupHandlers(mux)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
