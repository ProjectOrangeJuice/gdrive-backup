package nextcloud

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileList(t *testing.T) {

	c, err := NewClient()
	require.NoError(t, err)
	fileList, err := c.ListAllFiles("/google-test")
	require.NoError(t, err)

	for _, file := range fileList {
		t.Logf("File>%s, path>%s", file.Name(), file.Path)
	}

	t.Fail()
}
