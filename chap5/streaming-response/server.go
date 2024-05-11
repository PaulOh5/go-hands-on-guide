package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func longRunningProcess(logWriter *io.PipeWriter) {
	for i := 0; i <= 20; i++ {
		fmt.Fprintf(
			logWriter,
			`{"id": %d, "user_ip": "123.132.123.123", "event": "click_on_add_cart"}`,
			i,
		)
		fmt.Fprintln(logWriter)
		time.Sleep(1 * time.Second)
	}
	logWriter.Close()
}

func longRunningProcessHandler(w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	logReader, logWriter := io.Pipe()
	go longRunningProcess(logWriter)
	go progressStreamer(logReader, w, done)

	<-done
}

func progressStreamer(logReader *io.PipeReader, w http.ResponseWriter, done chan struct{}) {
	buf := make([]byte, 500)
	f, flushSupported := w.(http.Flusher)
	defer logReader.Close()
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	for {
		n, err := logReader.Read(buf)
		if err == io.EOF {
			break
		}
		w.Write(buf[:n])
		if flushSupported {
			f.Flush()
		}
	}

	done <- struct{}{}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/job", longRunningProcessHandler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
