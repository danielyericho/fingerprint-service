package main

import (
	"flag"
	"log"
	"net/http"

	"fingerprint-service/internal/api"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()
	srv, err := api.NewServer(nil)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}
	defer srv.Close()
	log.Printf("fingerprint-service listening on %s", *addr)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
