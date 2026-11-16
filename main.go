package main

import (
	"fmt"
	"log"
	"net/http"
)

func readiness(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	_, err := writer.Write([]byte("OK"))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", readiness)
	var server http.Server
	server.Handler = mux
	server.Addr = ":8080"
	log.Fatal(server.ListenAndServe())
}
