package backup

import (
	"fmt"
	"log"
	"time"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/gdrive"
	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/nextcloud"
)

type Item struct {
	Path             string
	ModificationTime time.Time
	Dir              bool
}

func GenerateFileListFromGoogle(gclient *gdrive.Client) ([]Item, error) {
	files, err := gclient.ListFiles()
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, file := range files {
		filePath := gclient.GetFullPath(file.Parents[0])
		parsedTime, err := time.Parse(time.RFC3339, file.ModifiedTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time for %s, %s", filePath, err)
		}
		items = append(items, Item{
			Path:             filePath + "/" + file.Name,
			ModificationTime: parsedTime,
			Dir:              file.MimeType == "application/vnd.google-apps.folder",
		})
	}

	return items, nil
}

func GenerateFileListFromNextcloud(nc *nextcloud.Client, dirs []string) (map[string][]Item, error) {
	fileList := make(map[string][]Item)
	for _, dir := range dirs {
		log.Printf("Searching %s", dir)
		files, err := nc.ListAllFiles(dir)
		if err != nil {
			return nil, err
		}
		fl := make([]Item, len(files))
		for index, file := range files {
			fl[index] = Item{
				Path:             file.Path,
				ModificationTime: file.ModTime(),
				Dir:              file.IsDir(),
			}
		}
		fileList[dir] = fl
	}
	return fileList, nil
}
