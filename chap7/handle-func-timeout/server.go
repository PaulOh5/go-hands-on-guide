package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handleUserAPI(w http.ResponseWriter, r *http.Request) {
	log.Println("I stared processing the request")
	time.Sleep(15 * time.Second)
	fmt.Fprint(w, "Hello, world!")
	log.Println("I finished processing the request")
}

func main() {
	timeoutDuration := 14 * time.Second

	userHandler := http.HandlerFunc(handleUserAPI)
	hTimeout := http.TimeoutHandler(
		userHandler,
		timeoutDuration,
		"I ran out of time\n",
	)

	mux := http.NewServeMux()
	mux.Handle("/api/users/", hTimeout)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
