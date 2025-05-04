package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

func main() {

	config := agent.Config{
		ServerURL:      "localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	var pollSec, reportSec string
	pflag.StringVarP(&config.ServerURL, "address", "a", config.ServerURL, "HTTP server endpoint address")
	pflag.StringVarP(&pollSec, "poll-interval", "p", "2", "Poll interval in seconds (default: 2)")
	pflag.StringVarP(&reportSec, "report-interval", "r", "10", "Report interval in seconds (default: 10)")
	pflag.BoolP("help", "h", false, "Show help message")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -a http://server:8080 -p 5 -r 30\n", os.Args[0])
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

	var err error
	if config.PollInterval, err = parseSeconds(pollSec); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid poll interval: %v\n\n", err)
		pflag.Usage()
		os.Exit(1)
	}

	if config.ReportInterval, err = parseSeconds(reportSec); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid report interval: %v\n\n", err)
		pflag.Usage()
		os.Exit(1)
	}

	if len(pflag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown arguments: %v\n\n", pflag.Args())
		pflag.Usage()
		os.Exit(1)
	}
	config.ServerURL = "http://" + config.ServerURL

	agent := agent.NewAgent(config)
	log.Printf("Starting agent with config:\n"+
		"  Server URL:      %s\n"+
		"  Poll Interval:   %v (%.0f seconds)\n"+
		"  Report Interval: %v (%.0f seconds)",
		config.ServerURL,
		config.PollInterval, config.PollInterval.Seconds(),
		config.ReportInterval, config.ReportInterval.Seconds())
	agent.Run()
}
