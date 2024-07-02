package backup

import (
	"time"
)

type Item struct {
	Path             string
	ModificationTime time.Time
}

// func GenerateFileListFromGoogle(files []*drive.File) []Item {
// 	var items []Item
// 	for _, file := range files {
// 		items = append(items, Item{
// 			Path:             file.,
// 			ModificationTime: file.ModifiedTime,
// 		})
// 	}
// }
