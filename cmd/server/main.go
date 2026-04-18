package main

import (
	"flag"
	"github.com/ccy/devices-monitor/internal/server"
	"github.com/ccy/devices-monitor/pkg/auth"
	"github.com/ccy/devices-monitor/pkg/config"
	"github.com/ccy/devices-monitor/pkg/storage"
	"log"
)

func main() {
	configFile := flag.String("config", "", "Configuration file path")
	addr := flag.String("addr", ":8080", "Server address")
	tlsAddr := flag.String("tls-addr", "", "TLS server address (optional)")
	certFile := flag.String("cert", "", "TLS certificate file (required for TLS)")
	keyFile := flag.String("key", "", "TLS key file (required for TLS)")
	secret := flag.String("secret", "your-secret-key-change-in-production", "JWT secret key")
	flag.Parse()

	var cfg *config.ServerConfig
	var err error

	if *configFile != "" {
		cfg, err = config.LoadServerConfig(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else {
		cfg = &config.ServerConfig{}
		if *addr != ":8080" {
			cfg.Addr = *addr
		}
		if *tlsAddr != "" {
			cfg.TLSAddr = *tlsAddr
		}
		if *certFile != "" {
			cfg.CertFile = *certFile
		}
		if *keyFile != "" {
			cfg.KeyFile = *keyFile
		}
		if *secret != "your-secret-key-change-in-production" {
			cfg.Secret = *secret
		}
	}

	st := storage.NewStorage()
	au := auth.NewAuth(cfg.Secret)
	srv := server.NewServer(st, au)

	if cfg.TLSAddr != "" {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			log.Fatal("TLS requires both cert and key options in config")
		}
		log.Printf("Starting TLS server on %s", cfg.TLSAddr)
		if err := srv.StartTLS(cfg.TLSAddr, cfg.CertFile, cfg.KeyFile); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Starting server on %s", cfg.Addr)
		if err := srv.Start(cfg.Addr); err != nil {
			log.Fatal(err)
		}
	}
}
