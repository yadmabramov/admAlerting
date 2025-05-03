package main

import (
	"log"

	"github.com/yadmabramov/admAlerting/internal/server"
)

func main() {
	srv := server.NewServer(":8080")
	log.Println("Server starting on :8080")
	log.Fatal(srv.ListenAndServe())
}
