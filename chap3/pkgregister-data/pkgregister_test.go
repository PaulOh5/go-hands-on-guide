package pkgregister

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func packageRegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		d := pkgRegisterResult{}
		err := r.ParseMultipartForm(5000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mForm := r.MultipartForm
		f := mForm.File["filedata"][0]
		d.Id = fmt.Sprintf(
			"%s-%s", mForm.Value["name"][0], mForm.Value["version"][0],
		)
		d.Filename = f.Filename
		d.Size = f.Size
		jsonData, err := json.Marshal(d)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(jsonData))
	} else {
		http.Error(
			w, "Invalid HTTP method specified", http.StatusMethodNotAllowed,
		)
		return
	}
}

func startTestPackageServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(packageRegHandler))
	return ts
}

func TestRegisterPackageData(t *testing.T) {
	ts := startTestPackageServer()
	defer ts.Close()
	p := pkgData{
		Name:     "mypackage",
		Version:  "0.1",
		Filename: "mypackage-0.1.tar.gz",
		Bytes:    strings.NewReader("some data"),
	}

	pResult, err := registerPackageData(ts.URL, p)
	if err != nil {
		t.Fatal(err)
	}
	if pResult.Id != fmt.Sprintf("%s-%s", p.Name, p.Version) {
		t.Errorf("Expected id to be %s-%s, Got: %s", p.Name, p.Version, pResult.Id)
	}
	if pResult.Filename != p.Filename {
		t.Errorf("Expected filename to be %s, Got: %s", p.Filename, pResult.Filename)
	}
	if pResult.Size != int64(len("some data")) {
		t.Errorf("Expected size to be %d, Got: %d", len("some data"), pResult.Size)
	}
}
