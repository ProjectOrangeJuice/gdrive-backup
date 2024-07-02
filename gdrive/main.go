package main

import (
	"flag"
	"log"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/config"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/gdrive"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/nextcloud"
)

var tokenFlag string

func main() {
	flag.StringVar(&tokenFlag, "auth", "", "Auth token")
	flag.Parse()

	// Read config json
	conf := config.ReadConfig("../config.json")
	// Setup gdrive..
	g, err := gdrive.NewClient(tokenFlag, conf.GoogleBaseFolder)
	if err != nil {
		log.Fatalf("Could not setup google drive because %s", err)
	}

	// Setup nextcloud
	nc, err := nextcloud.NewClient()
	if err != nil {
		log.Fatalf("Could not setup nextcloud because %s", err)
	}

	// Generate the list of files from google, with their modification times
	googleFiles, err := g.ListFiles()

	// Generate the list of files from nextcloud, with their modification times

	// Compare the list of files to work out what needs to be uploaded
	// Generate a list of files, use works to upload the files
}
