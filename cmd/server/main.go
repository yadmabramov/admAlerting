package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/pflag"
	"github.com/yadmabramov/admAlerting/internal/server"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if sec, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(sec) * time.Second
		}
	}
	return defaultValue
}

func validateAndNormalizeServerURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid server URL: %v", err)
	}

	if u.Port() == "" {
		u.Host = u.Hostname() + ":8080"
	}

	return u.String(), nil
}

func main() {
	defaultConfig := server.Config{
		Addr:          "localhost:8080",
		StoreInterval: 5 * time.Second,
		StoragePath:   "metrics-db.json",
		Restore:       true,
		DatabaseDSN:   "",
	}

	config := server.Config{
		Addr:          getEnv("ADDRESS", defaultConfig.Addr),
		StoreInterval: getEnvDuration("STORE_INTERVAL", defaultConfig.StoreInterval),
		StoragePath:   getEnv("FILE_STORAGE_PATH", defaultConfig.StoragePath),
		Restore:       getEnvBool("RESTORE", defaultConfig.Restore),
		DatabaseDSN:   getEnv("DATABASE_DSN", defaultConfig.DatabaseDSN),
	}

	var flagAddr, flagStoreInt, flagStoragePath, flagDatabaseDSN string
	var flagRestore bool
	pflag.StringVarP(&flagAddr, "address", "a", "", "HTTP server endpoint address (env: ADDRESS)")
	pflag.StringVarP(&flagStoreInt, "store-interval", "i", "", "Interval to save metrics to disk in seconds (env: STORE_INTERVAL)")
	pflag.StringVarP(&flagStoragePath, "file-storage-path", "f", "", "Path to file for saving metrics (env: FILE_STORAGE_PATH)")
	pflag.StringVarP(&flagDatabaseDSN, "database-dsn", "d", "", "Database connection string (env: DATABASE_DSN)") г
	pflag.BoolVarP(&flagRestore, "restore", "r", true, "Restore metrics from file (env: RESTORE)")
	pflag.BoolP("help", "h", false, "Show help message")
	pflag.BoolP("version", "v", false, "Show version information")
	pflag.CommandLine.SortFlags = false

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  ADDRESS            HTTP server endpoint address (highest priority)\n")
		fmt.Fprintf(os.Stderr, "  STORE_INTERVAL     Interval to save metrics to disk in seconds\n")
		fmt.Fprintf(os.Stderr, "  FILE_STORAGE_PATH  Path to file for saving metrics\n")
		fmt.Fprintf(os.Stderr, "  RESTORE            Restore metrics from file (true/false)\n")
		fmt.Fprintf(os.Stderr, "\nPriority: ENV > FLAGS > DEFAULTS\n")
	}

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		pflag.Usage()
		os.Exit(1)
	}

	if flagAddr != "" && os.Getenv("ADDRESS") == "" {
		config.Addr = flagAddr
	}
	if flagStoreInt != "" && os.Getenv("STORE_INTERVAL") == "" {
		if interval, err := strconv.ParseInt(flagStoreInt, 10, 64); err == nil {
			config.StoreInterval = time.Duration(interval) * time.Second
		}
	}
	if flagStoragePath != "" && os.Getenv("FILE_STORAGE_PATH") == "" {
		config.StoragePath = flagStoragePath
	}
	if pflag.Lookup("restore").Changed && os.Getenv("RESTORE") == "" {
		config.Restore = flagRestore
	}
	if flagDatabaseDSN != "" && os.Getenv("DATABASE_DSN") == "" {
		config.DatabaseDSN = flagDatabaseDSN
	}
	normalizedURL, err := validateAndNormalizeServerURL(config.Addr)
	if err != nil {
		log.Fatalf("Server URL validation failed: %v", err)
	}
	config.Addr = normalizedURL

	srv := server.NewServer(config)
	log.Printf("Server starting on %s", config.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
