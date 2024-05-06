package pkgregister

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func packageRegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		p := pkgData{}
		d := pkgRegisterResult{}
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = json.Unmarshal(data, &p)
		if err != nil || len(p.Name) == 0 || len(p.Version) == 0 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		d.Id = p.Name + "-" + p.Version
		jsonData, err := json.Marshal(d)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonData))
	} else {
		http.Error(w, "Invalid HTTP method specified", http.StatusMethodNotAllowed)
		return
	}
}

func StartTestPackageServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(packageRegHandler))
	return ts
}

func TestRegisterPackageData(t *testing.T) {
	ts := StartTestPackageServer()
	defer ts.Close()
	p := pkgData{}
	resp, err := registerPackageData(ts.URL, p)
	if err == nil {
		t.Fatal("Expected error to be non nil, got nil")
	}
	if len(resp.Id) != 0 {
		t.Errorf("Expected package id to be empty, got %s", resp.Id)
	}
}
