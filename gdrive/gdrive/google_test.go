package gdrive

import (
	"log"
	"testing"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/config"
	"github.com/stretchr/testify/require"
)

func TestGoogleList(t *testing.T) {
	conf := config.ReadConfig("../../config.json")
	client, err := NewClient("", conf.GoogleBaseFolder)
	require.NoError(t, err)
	files, err := client.ListFiles()
	require.NoError(t, err)

	for _, f := range files {
		fullPath := GetFullPath(client, f.Parents[0])
		log.Printf("file: %s,  path %s ", f.Name, fullPath)
	}
	t.Fail()
}
