package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func packageHTTPHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "package1-0.1")
	case "POST":
		p := pkgData{}
		d := pkgRegisterResult{}
		defer r.Body.Close()
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
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
			d.ID = p.Name + "-" + p.Version
			jsonData, err := json.Marshal(d)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, string(jsonData))
		} else if strings.Contains(contentType, "multipart/form-data") {
			err := r.ParseMultipartForm(5000)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			mForm := r.MultipartForm
			f := mForm.File["filedata"][0]
			d.ID = fmt.Sprintf(
				"%s-%s", mForm.Value["name"][0], mForm.Value["version"][0],
			)
			d.Name = f.Filename
			d.Size = f.Size
			jsonData, err := json.Marshal(d)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, string(jsonData))
		} else {
			errorMsg := fmt.Sprintf("Invalid Content-type: %v", r.Header.Get("Content-Type"))
			http.Error(w, errorMsg, http.StatusBadRequest)
		}
	default:
		http.Error(w, "Invalid HTTP method specified", http.StatusMethodNotAllowed)
	}
}

func StartTestPackageServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(packageHTTPHandler))
	return ts
}
