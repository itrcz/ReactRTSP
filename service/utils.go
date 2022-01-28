package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"strings"
)

func createUUID() (uuid string) {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func createURL(request *http.Request) (url string) {
	var urlArray []string

	protocol := request.FormValue("protocol")
	login := request.FormValue("login")
	password := request.FormValue("password")
	host := request.FormValue("host")
	port := request.FormValue("port")
	uri := request.FormValue("uri")

	urlArray = append(urlArray, protocol, "://")
	if len(login) > 0 {
		urlArray = append(urlArray, login, ":", password, "@")
	}
	urlArray = append(urlArray, host, ":", port, uri)
	return strings.Join(urlArray, "")
}
