package backup

import "log"

type Config struct {
	Directories []string
}

// go through the lists and find out what files have changed or are missing
func FindChanges(nextcloudList []Item, googleList []Item) []Item {
	var uploadList []Item
	for _, nextcloudItem := range nextcloudList {
		if nextcloudItem.Dir {
			continue
		}
		uploadFile := true
		for _, googleItem := range googleList {
			if nextcloudItem.Path == googleItem.Path {
				if nextcloudItem.ModificationTime == googleItem.ModificationTime {
					log.Printf("File %s has not changed", nextcloudItem.Path)
					uploadFile = false
					break
				}
			}
		}

		if uploadFile {
			uploadList = append(uploadList, nextcloudItem)
		}
	}

	return uploadList
}
