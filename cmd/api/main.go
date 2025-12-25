package main

import (
	"log"
	"net/http"

	"highlightiq-server/internal/http/router"
)

func main() {
	r := router.New()

	addr := ":8080"
	log.Println("API listening on", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
