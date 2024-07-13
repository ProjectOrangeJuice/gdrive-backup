package backup

import (
	"fmt"
	"log"
	"time"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/config"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/gdrive"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/nextcloud"
)

type Item struct {
	Path             string
	ModificationTime time.Time
	Dir              bool
	Name             string
}

func GenerateFileListFromGoogle(gclient *gdrive.Client) ([]Item, error) {
	files, err := gclient.ListFiles()
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, file := range files {
		var filePath string
		var err error
		if file.MimeType != "application/vnd.google-apps.folder" {
			filePath, err = gclient.GetFullPath(file.Parents[0])
			if err != nil {
				return nil, fmt.Errorf("when getting the full path for %s, got error %s", file.Name, err)
			}
		}

		parsedTime, err := time.Parse(time.RFC3339, file.ModifiedTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time for %s, %s", filePath, err)
		}
		items = append(items, Item{
			Path:             filePath + "/" + file.Name,
			ModificationTime: parsedTime,
			Dir:              file.MimeType == "application/vnd.google-apps.folder",
			Name:             file.Name,
		})
	}

	return items, nil
}

func GenerateFileListFromNextcloud(nc *nextcloud.Client, dirs []config.DirectoryConfig) (map[string][]Item, error) {
	fileList := make(map[string][]Item)
	for _, dir := range dirs {
		log.Printf("Searching %s", dir)
		files, err := nc.ListAllFiles(dir.Dir)
		if err != nil {
			return nil, err
		}
		fl := make([]Item, len(files))
		for index, file := range files {
			fl[index] = Item{
				Name:             file.Name(),
				Path:             file.Path,
				ModificationTime: file.ModTime(),
				Dir:              file.IsDir(),
			}
		}
		fileList[dir.Dir] = fl
	}
	return fileList, nil
}
