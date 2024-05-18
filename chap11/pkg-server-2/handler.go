package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

func getOwnerId() int {
	return rand.Intn(4) + 1
}

func packageRegHandler(
	w http.ResponseWriter,
	r *http.Request,
	config appConfig,
) {
	d := pkgRegisterResponse{}
	err := r.ParseMultipartForm(5000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mForm := r.MultipartForm
	fHeader := mForm.File["filedata"][0]

	packageName := mForm.Value["name"][0]
	packageVersion := mForm.Value["version"][0]
	packageOwner := getOwnerId()

	q := pkgQueryParams{
		ownerId: packageOwner,
		version: packageName,
		name:    packageName,
	}
	pkgResults, err := queryDb(config, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(pkgResults) != 0 {
		http.Error(w, "Package version for the owner exists", http.StatusBadRequest)
		return
	}

	d.ID = fmt.Sprintf(
		"%d/%s-%s-%s",
		packageOwner,
		packageName,
		packageVersion,
		fHeader.Filename,
	)
	nBytes, err := uploadData(config, d.ID, fHeader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = updateDb(
		config,
		pkgRow{
			OwnerId:       packageOwner,
			Name:          packageName,
			Version:       packageVersion,
			ObjectStoreId: d.ID,
		},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	config.logger.Printf(
		"Package uploaded: %s. Bytes written: %d\n",
		d.ID,
		nBytes,
	)
	jsonData, err := json.Marshal(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonData))
}

func packageGetHandler(
	w http.ResponseWriter,
	r *http.Request,
	config appConfig,
) {
	q, err := parseQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if q.ownerId == -1 || len(q.name) == 0 || len(q.version) == 0 {
		http.Error(w, "Must specify package owner, name and version", http.StatusBadRequest)
		return
	}

	pkgResults, err := queryDb(config, q)
	log.Println(pkgResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(pkgResults) == 0 {
		http.Error(w, "No package found", http.StatusNotFound)
		return
	}

	packageID := pkgResults[0].ObjectStoreId

	exists, err := config.packageBucket.Exists(r.Context(), packageID)
	if err != nil || !exists {
		http.Error(w, "invalid package ID", http.StatusNotFound)
		return
	}

	download := r.URL.Query().Get("download")
	if download == "true" {
		reader, err := config.packageBucket.NewReader(r.Context(), packageID, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+packageID+"\"")
		if _, err := io.Copy(w, reader); err != nil {
			http.Error(w, "Failed to send file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		url, err := config.packageBucket.SignedURL(r.Context(), packageID, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func packageQueryHandler(
	w http.ResponseWriter,
	r *http.Request,
	config appConfig,
) {
	q, err := parseQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	pkgResults, err := queryDb(config, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(pkgResults)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(jsonData))
}

func packageHandler(
	w http.ResponseWriter,
	r *http.Request,
	config appConfig,
) {
	switch r.Method {
	case "GET":
		packageQueryHandler(w, r, config)
	case "POST":
		packageRegHandler(w, r, config)
	default:
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
	}
}

func parseQuery(r *http.Request) (pkgQueryParams, error) {
	queryParams := r.URL.Query()
	owner := queryParams.Get("owner_id")
	if len(owner) == 0 {
		owner = "-1"
	}
	ownerId, err := strconv.Atoi(owner)
	if err != nil {
		return pkgQueryParams{}, err
	}
	name := queryParams.Get("name")
	version := queryParams.Get("version")

	q := pkgQueryParams{
		ownerId: ownerId,
		name:    name,
		version: version,
	}
	return q, nil
}
