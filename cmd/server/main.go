package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
	"github.com/yadmabramov/admAlerting/internal/server"
)

func main() {

	var (
		addr    string
		help    bool
		version bool
	)

	pflag.StringVarP(&addr, "address", "a", "localhost:8080", "HTTP server endpoint address")
	pflag.BoolVarP(&help, "help", "h", false, "Show help message")
	pflag.BoolVarP(&version, "version", "v", false, "Show version information")
	pflag.CommandLine.SortFlags = false

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		pflag.PrintDefaults()
	}

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		pflag.Usage()
		os.Exit(1)
	}

	if help {
		pflag.Usage()
		return
	}

	if version {
		fmt.Println("admAlerting v1.0.0")
		return
	}

	if len(pflag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown arguments: %v\n", pflag.Args())
		pflag.Usage()
		os.Exit(1)
	}

	srv := server.NewServer(addr)
	log.Printf("Server starting on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
