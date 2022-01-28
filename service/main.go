package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/websocket"
)

func main() {
	log.Println("Ready")
	http.Handle("/stream", websocket.Handler(ws))
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":3333", nil)
}
