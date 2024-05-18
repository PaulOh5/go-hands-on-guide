package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPackageGetHandler(t *testing.T) {
	packageBucket, err := getTestBucket(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer packageBucket.Close()

	testObjectId := "pkg-0.1-pkg-0.1.tar.gz"
	err = packageBucket.WriteAll(
		context.Background(),
		testObjectId, []byte("test-data"),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	testC, testDb, err := getTestDb()
	if err != nil {
		t.Fatal(err)
	}
	defer testC.Terminate(context.Background())

	config := appConfig{
		logger: log.New(
			os.Stdout, "",
			log.Ldate|log.Ltime|log.Lshortfile,
		),
		packageBucket: packageBucket,
		db:            testDb,
	}

	err = updateDb(
		config,
		pkgRow{
			OwnerId:       1,
			Name:          "pkg",
			Version:       "0.1",
			ObjectStoreId: testObjectId,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	setupHandlers(mux, config)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	var redirectUrl string
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectUrl = req.URL.String()
			return errors.New("no redirect")
		},
	}

	_, err = client.Get(
		"http://localhost:8080/api/packages/download?owner_id=1&name=pkg&version=0.1",
	)
	if err == nil {
		t.Fatal("Expected error: no redirect, but Got: nil")
	}
	if !strings.HasPrefix(redirectUrl, "file:///") {
		t.Fatalf("Expected redirect url to start with file:///, got: %v", redirectUrl)
	}

}
