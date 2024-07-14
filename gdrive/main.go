package main

import (
	"flag"
	"log"
	"sync"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/backup"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/config"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/gdrive"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/nextcloud"
)

var (
	tokenFlag string
	dryRun    bool
)

func main() {
	flag.StringVar(&tokenFlag, "auth", "", "Auth token")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run")
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
			if !dryRun {
				uploadChanges(changes, nc, g, getEncFromConfig(key, conf.Directories), 4)
			}
		} else {
			log.Printf("No changes")
		}
	}
}

// my dodgy code means I dont have the enc key when I want it
func getEncFromConfig(key string, config []config.DirectoryConfig) string {
	for _, dir := range config {
		if dir.Dir == key {
			return dir.Encryption
		}
	}
	log.Panicf("Could not get key for %s", key)
	return ""
}

func uploadChanges(changes []backup.Item, nc *nextcloud.Client, g *gdrive.Client, encryption string, numWorkers int) {
	log.Printf("Uploading changes with %d workers", numWorkers)

	// Create a channel to receive upload tasks
	tasks := make(chan backup.Item, len(changes))

	// Create a wait group to track worker completion
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	// Start the workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for change := range tasks {
				f, err := nc.DownloadFile(change.Path)
				if err != nil {
					log.Printf("Failed to get file for download: %s", err)
					continue // Skip to the next file
				}

				if encryption != "" {
					f, err = backup.Encrypt([]byte(encryption), f)
					if err != nil {
						log.Panicf("Failed to encrypt file: %s", err)
					}
				}

				gfile := gdrive.File{
					Name:         change.Name,
					Path:         change.Path,
					Reader:       f,
					ModifiedTime: change.ModificationTime,
				}

				err = g.UploadFile(gfile)
				f.Close()
				if err != nil {
					log.Printf("Failed to upload file: %s", err)
					continue // Skip to the next file
				}
				log.Printf("Uploaded %s", change.Name)
			}
		}()
	}

	// Send the upload tasks to the channel
	for _, change := range changes {
		tasks <- change
	}
	close(tasks)

	// Wait for all workers to finish
	wg.Wait()
}
