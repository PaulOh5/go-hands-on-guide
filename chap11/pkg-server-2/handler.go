package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	queryParams := r.URL.Query()
	owner := queryParams.Get("owner_id")
	name := queryParams.Get("name")
	version := queryParams.Get("version")
	download := queryParams.Get("download")

	if len(owner) == 0 || len(name) == 0 || len(version) == 0 {
		http.Error(w, "Must specify package owner, name and version", http.StatusBadRequest)
		return
	}
	ownerId, err := strconv.Atoi(owner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	q := pkgQueryParams{
		ownerId: ownerId,
		version: version,
		name:    name,
	}
	pkgResults, err := queryDb(config, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if len(pkgResults) == 0 {
		http.Error(w, "No package found", http.StatusNotFound)
	}

	packageID := pkgResults[0].ObjectStoreId

	exists, err := config.packageBucket.Exists(r.Context(), packageID)
	if err != nil || !exists {
		http.Error(w, "invalid package ID", http.StatusNotFound)
		return
	}

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

func packageHandler(
	w http.ResponseWriter,
	r *http.Request,
	config appConfig,
) {
	switch r.Method {
	case "GET":
		packageGetHandler(w, r, config)
	case "POST":
		packageRegHandler(w, r, config)
	default:
		http.Error(w, "Invalid Method", http.StatusMethodNotAllowed)
	}
}
