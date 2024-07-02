package config

import (
	"encoding/json"
	"log"
	"os"
)

type config struct {
	Directories      []string `json:"directories"`
	GoogleBaseFolder string   `json:"googleBaseFolder"`
}

func ReadConfig(dir string) config {
	f, err := os.Open(dir)
	if err != nil {
		log.Fatalf("could not read config, %s", err)
	}
	defer f.Close()
	var c config
	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		log.Fatalf("could not read config, %s", err)
	}
	return c
}
