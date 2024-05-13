package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	var buf bytes.Buffer
	mux := http.NewServeMux()
	l := log.New(&buf, "test-tls-server", log.Lshortfile|log.LstdFlags)
	m := setupHandlersAndMiddleWare(mux, l)

	ts := httptest.NewUnstartedServer(m)
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	client := ts.Client()
	_, err := client.Get(ts.URL + "/api")
	if err != nil {
		t.Fatal(err)
	}

	expected := "protocoal=HTTP/2.0 path=/api method=GET duration="
	mLogs := buf.String()
	if !strings.Contains(mLogs, expected) {
		t.Fatalf(
			"Expected logs to contain %s, Found: %s\n",
			expected, mLogs,
		)
	}
}
