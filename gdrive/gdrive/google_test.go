package gdrive

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoogleList(t *testing.T) {
	client, err := NewClient("")
	require.NoError(t, err)
	files, err := client.ListFiles()
	require.NoError(t, err)

	for _, f := range files {
		log.Printf("file: %s, ext %s ", f.Name, f.FileExtension)
	}
	t.Fail()
}
