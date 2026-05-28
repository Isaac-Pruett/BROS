package main

import (
	"fmt"
	"os"

	zenoh "github.com/eclipse-zenoh/zenoh-go/zenoh"
)

func cfgFromEnv() (zenoh.Config, error) {
	configPath := os.Getenv("ZENOH_CONFIG")

	var cfg zenoh.Config
	var err error

	if configPath != "" {
		cfg, err = zenoh.NewConfigFromFile(configPath)
	} else {
		cfg = zenoh.NewConfigDefault()
	}

	return cfg, err
}

func main() {

	cfg, err := cfgFromEnv()
	if err != nil {
		cfg = zenoh.NewConfigDefault()
	}

	fmt.Println("Opening session...")
	session, err := zenoh.Open(cfg, nil)

	if err != nil {
		fmt.Printf("failed to open zenoh session: %v\n", err)
		os.Exit(1)
	}

	defer session.Drop()

	fmt.Println("session opened")
}
