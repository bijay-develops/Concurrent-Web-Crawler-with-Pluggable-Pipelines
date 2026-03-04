package main

import (
	"log"
	"net/http"
	"os"

	"crawler/internal/httpapi"
)

func main() {
	mux := http.NewServeMux()

	h := httpapi.NewHandler()
	h.Register(mux)

	addr := ":8090"
	if v := os.Getenv("API_PORT"); v != "" {
		addr = ":" + v
	}

	log.Printf("API listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
