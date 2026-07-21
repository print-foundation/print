package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/print-foundation/print/internal/builder"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/internal/network"
)

func main() {
	log := logging.New(logging.Options{Level: "info"})

	if len(os.Args) > 1 && os.Args[1] == "autoinstall" {
		if err := runAutoinstall(log); err != nil {
			log.Error("autoinstall failed", "error", err)
			fmt.Fprintln(os.Stderr, "print-client:", err)
			os.Exit(1)
		}
		return
	}

	if err := prepare(log); err != nil {
		log.Error("client failed", "error", err)
		fmt.Fprintln(os.Stderr, "print-client:", err)
		os.Exit(1)
	}
}

func prepare(log logging.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	log.Info("Pr!nt cloud environment preparing", "distro", cfg.Distro, "mirror", cfg.Mirror)

	net := network.New(log)
	if cfg.WiFi.SSID != "" {
		log.Info("connecting to wifi", "ssid", cfg.WiFi.SSID)
		if err := net.ConnectWiFi(context.Background(), "", cfg.WiFi.SSID, cfg.WiFi.Passphrase); err != nil {
			return fmt.Errorf("wifi connect: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if !network.NewConnectivity().Online(ctx) {
		return fmt.Errorf("no internet connectivity; cannot prepare cloud environment")
	}
	log.Info("online; mirror reachable", "mirror", cfg.Mirror)

	seed, err := generateSeed(cfg)
	if err != nil {
		return fmt.Errorf("generate seed: %w", err)
	}
	if err := os.WriteFile(seedPath, []byte(seed), 0o644); err != nil {
		return fmt.Errorf("write seed: %w", err)
	}
	log.Info("environment ready", "seed", seedPath, "hint", "run `print-client autoinstall` for unattended install")
	return nil
}

func loadConfig() (builder.ClientConfig, error) {
	cfgPath := os.Getenv("PRINT_CONFIG")
	if cfgPath == "" {
		cfgPath = "/print-client.json"
	}
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		return builder.ClientConfig{}, fmt.Errorf("read config: %w", err)
	}
	var cfg builder.ClientConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return builder.ClientConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

const seedPath = "/print-seed.cfg"

var _ = flag.Bool
