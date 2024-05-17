package main

import "io"

type pkgData struct {
	Name     string
	Version  string
	Filename string
	Bytes    io.Reader
}

type pkgRegisterResponse struct {
	ID string `json:"id"`
}

type pkgQueryParams struct {
	name    string
	version string
	ownerId int
}

type pkgRow struct {
	OwnerId       int
	Name          string
	Version       string
	ObjectStoreId string
	Created       string
}
