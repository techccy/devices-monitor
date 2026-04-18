package main

import (
	"flag"
	"github.com/ccy/devices-monitor/internal/agent"
	"github.com/ccy/devices-monitor/pkg/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("config", "", "Configuration file path")
	serverURL := flag.String("server", "ws://localhost:8080/api/ws", "Server WebSocket URL")
	deviceID := flag.String("id", "", "Device ID (optional)")
	deviceKey := flag.String("key", "", "Device key (optional)")
	heartbeat := flag.Int("heartbeat", 300, "Heartbeat interval in seconds")
	flag.Parse()

	var cfg *config.AgentConfig
	var err error

	if *configFile != "" {
		cfg, err = config.LoadAgentConfig(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else {
		cfg = &config.AgentConfig{}
		if *serverURL != "ws://localhost:8080/api/ws" {
			cfg.ServerURL = *serverURL
		}
		if *deviceID != "" {
			cfg.DeviceID = *deviceID
		}
		if *deviceKey != "" {
			cfg.DeviceKey = *deviceKey
		}
		if *heartbeat != 300 {
			cfg.Heartbeat = *heartbeat
		}
	}

	ag := agent.NewAgent(cfg.ServerURL, cfg.DeviceID, cfg.DeviceKey, cfg.Heartbeat)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down agent...")
		ag.Stop()
	}()

	if err := ag.Start(); err != nil {
		log.Fatal(err)
	}
}
