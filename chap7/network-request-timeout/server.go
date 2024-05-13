package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	log.Println("ping: Got a request")
	fmt.Fprintf(w, "pong")
}

func doSomeOtherWork() {
	time.Sleep(2 * time.Second)
}

func handleUserAPI(w http.ResponseWriter, r *http.Request) {
	log.Println("I started processing the request")

	doSomeOtherWork()

	log.Println("Outgoing HTTP request")

	req, err := http.NewRequestWithContext(
		r.Context(),
		"GET",
		"http://localhost:8080/ping", nil,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	trace := &httptrace.ClientTrace{
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			log.Printf("DNS Info: %+v\n", dnsInfo)
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			log.Printf("Got Conn: %+v\n", connInfo)
		},
		ConnectStart: func(network, addr string) {
			log.Printf("Connect Start, network=%s addr=%s", network, addr)
		},
		WroteRequest: func(wri httptrace.WroteRequestInfo) {
			log.Printf("Wrote Request: %+v\n", wri)
		},
	}

	ctxTrace := httptrace.WithClientTrace(req.Context(), trace)
	req = req.WithContext(ctxTrace)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	fmt.Fprint(w, string(data))
	log.Println("I finished processing the request")
}

func main() {
	timeoutDuration := 3 * time.Second

	userHandler := http.HandlerFunc(handleUserAPI)
	hTimeout := http.TimeoutHandler(userHandler, timeoutDuration, "I ran out of time\n")

	mux := http.NewServeMux()
	mux.Handle("/api/users/", hTimeout)
	mux.HandleFunc("/ping", handlePing)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
