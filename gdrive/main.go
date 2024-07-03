package main

import (
	"flag"
	"log"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/backup"
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
	log.Printf("Connecting to google")
	g, err := gdrive.NewClient(tokenFlag, conf.GoogleBaseFolder)
	if err != nil {
		log.Fatalf("Could not setup google drive because %s", err)
	}

	// Setup nextcloud
	log.Printf("Connecting to nextcloud")
	nc, err := nextcloud.NewClient()
	if err != nil {
		log.Fatalf("Could not setup nextcloud because %s", err)
	}

	// Generate the list of files from google, with their modification times
	log.Printf("Searching google")
	googleFiles, err := backup.GenerateFileListFromGoogle(g)
	if err != nil {
		log.Fatalf("Could not generate google drive list, %s", err)
	}

	// Generate the list of files from nextcloud, with their modification times
	log.Printf("Searching nextcloud")
	nextcloudFiles, err := backup.GenerateFileListFromNextcloud(nc, conf.Directories)
	if err != nil {
		log.Fatalf("Could not generate nextcloud list, %s", err)
	}
	log.Printf("*** comparing changes ***")
	// Compare the list of files to work out what needs to be uploaded
	// Generate a list of files, use works to upload the files
	for key, value := range nextcloudFiles {
		log.Printf("Checking for changes in %s", key)
		changes := backup.FindChanges(value, googleFiles)
		if len(changes) > 0 {
			log.Printf("Found changes.. [%+v]", changes)
			uploadChanges(changes, nc, g)
		} else {
			log.Printf("No changes")
		}
	}
}

func uploadChanges(changes []backup.Item, nc *nextcloud.Client, g *gdrive.Client) {
	log.Printf("Uploading changes")

	for _, change := range changes {
		f, err := nc.DownloadFile(change.Path)
		if err != nil {
			log.Fatalf("Failed to get file for download, %s", err)
		}

		gfile := gdrive.File{
			Name:   change.Name,
			Path:   change.Path,
			Reader: f,
		}

		err = g.UploadFile(gfile)
		if err != nil {
			log.Fatalf("Failed to upload file, %s", err)
		}

	}

}
