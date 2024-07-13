package config

import (
	"encoding/json"
	"log"
	"os"
)

type config struct {
	Directories      []DirectoryConfig `json:"directories"`
	GoogleBaseFolder string            `json:"googleBaseFolder"`
}

type DirectoryConfig struct {
	Dir        string
	Encryption string
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
