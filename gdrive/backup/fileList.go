package backup

import (
	"fmt"
	"time"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/gdrive"
)

type Item struct {
	Path             string
	ModificationTime time.Time
}

func GenerateFileListFromGoogle(gclient gdrive.Client) ([]Item, error) {
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
		})
	}

	return items, nil
}
