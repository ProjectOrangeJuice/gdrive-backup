package nextcloud

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileList(t *testing.T) {

	c, err := NewClient()
	require.NoError(t, err)
	c.ListFiles()
	t.Fail()
}
