package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/pflag"
	"github.com/yadmabramov/admAlerting/internal/agent"
)

func parseSeconds(s string) (time.Duration, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("must be integer number of seconds")
	}
	return time.Duration(sec) * time.Second, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
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
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	// Парсим URL для проверки структуры
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

func main() {
	defaultConfig := agent.Config{
		ServerURL:      "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Key:            "",
		RateLimit:      1,
	}

	config := agent.Config{
		ServerURL:      getEnv("ADDRESS", defaultConfig.ServerURL),
		PollInterval:   getEnvDuration("POLL_INTERVAL", defaultConfig.PollInterval),
		ReportInterval: getEnvDuration("REPORT_INTERVAL", defaultConfig.ReportInterval),
		Key:            getEnv("KEY", defaultConfig.Key),
		RateLimit:      getEnvInt("RATE_LIMIT", defaultConfig.RateLimit),
	}

	var (
		flagAddress   string
		flagPoll      string
		flagReport    string
		flagKey       string
		flagRateLimit int
	)

	pflag.StringVarP(&flagAddress, "address", "a", "", "HTTP server endpoint address")
	pflag.StringVarP(&flagPoll, "poll-interval", "p", "", "Poll interval in seconds")
	pflag.StringVarP(&flagReport, "report-interval", "r", "", "Report interval in seconds")
	pflag.BoolP("help", "h", false, "Show help message")
	pflag.StringVarP(&flagKey, "key", "k", "", "Secret key for hash calculation (env: KEY)")
	pflag.IntVarP(&flagRateLimit, "limit", "l", 0, "Rate limit for outgoing requests (env: RATE_LIMIT)")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables (highest priority):\n")
		fmt.Fprintf(os.Stderr, "  ADDRESS          HTTP server endpoint address\n")
		fmt.Fprintf(os.Stderr, "  POLL_INTERVAL    Poll interval in seconds\n")
		fmt.Fprintf(os.Stderr, "  REPORT_INTERVAL  Report interval in seconds\n")
		fmt.Fprintf(os.Stderr, "  RATE_LIMIT       Rate limit for outgoing requests\n")
		fmt.Fprintf(os.Stderr, "\nPriority: ENV > FLAGS > DEFAULTS\n")
	}

	if err := pflag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		pflag.Usage()
		os.Exit(1)
	}

	if help, _ := pflag.CommandLine.GetBool("help"); help {
		pflag.Usage()
		return
	}

	if flagAddress != "" && os.Getenv("ADDRESS") == "" {
		config.ServerURL = flagAddress
	}
	if flagPoll != "" && os.Getenv("POLL_INTERVAL") == "" {
		if interval, err := parseSeconds(flagPoll); err == nil {
			config.PollInterval = interval
		} else {
			log.Fatalf("Invalid poll interval: %v", err)
		}
	}
	if flagReport != "" && os.Getenv("REPORT_INTERVAL") == "" {
		if interval, err := parseSeconds(flagReport); err == nil {
			config.ReportInterval = interval
		} else {
			log.Fatalf("Invalid report interval: %v", err)
		}
	}
	if flagKey != "" && os.Getenv("KEY") == "" {
		config.Key = flagKey
	}
	if flagRateLimit > 0 && os.Getenv("RATE_LIMIT") == "" {
		config.RateLimit = flagRateLimit
	}

	normalizedURL, err := validateAndNormalizeServerURL(config.ServerURL)
	if err != nil {
		log.Fatalf("Server URL validation failed: %v", err)
	}
	config.ServerURL = normalizedURL

	agent := agent.NewAgent(config)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		agent.Shutdown()
	}()

	agent.Run()
}
