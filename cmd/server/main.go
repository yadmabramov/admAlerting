package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/yadmabramov/admAlerting/internal/server"
)

func validateAndNormalizeServerURL(rawURL string) (string, error) {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %v", err)
	}

	if u.Port() == "" {
		if u.Scheme == "https" {
			u.Host = u.Hostname() + ":443"
		} else {
			u.Host = u.Hostname() + ":8080"
		}
	}

	if u.Hostname() == "" {
		return "", fmt.Errorf("server host cannot be empty")
	}

	return u.String(), nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	addr := getEnv("ADDRESS", "localhost:8080")

	var flagAddr string
	pflag.StringVarP(&flagAddr, "address", "a", "", "HTTP server endpoint address (env: ADDRESS)")
	pflag.BoolP("help", "h", false, "Show help message")
	pflag.BoolP("version", "v", false, "Show version information")
	pflag.CommandLine.SortFlags = false

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  ADDRESS          HTTP server endpoint address (highest priority)\n")
		fmt.Fprintf(os.Stderr, "\nPriority: ENV > FLAGS > DEFAULTS\n")
	}

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		pflag.Usage()
		os.Exit(1)
	}

	if flagAddr != "" && os.Getenv("ADDRESS") == "" {
		addr = flagAddr
	}

	normalizedURL, err := validateAndNormalizeServerURL(addr)
	if err != nil {
		log.Fatalf("Server URL validation failed: %v", err)
	}
	addr = normalizedURL

	srv := server.NewServer(addr)
	log.Printf("Server starting on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
