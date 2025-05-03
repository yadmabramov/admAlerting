package main

import (
	"log"
	"time"

	"github.com/yadmabramov/admAlerting/internal/agent"
)

func main() {
	config := agent.Config{
		ServerURL:      "http://localhost:8080",
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
	}

	agent := agent.NewAgent(config)
	log.Println("Agent started with config:", config)
	agent.Run()
}
