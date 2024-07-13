package backup

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncrypt(t *testing.T) {
	// Read the file into the reader
	file, err := os.Open("test/test.base")
	require.NoError(t, err)
	defer file.Close()

	// Encrypt the file
	key := []byte("PPKpKqSMGfX43h2qJbP9cpkn886u9Y2D")
	encFile, err := Encrypt(key, file)
	require.NoError(t, err)

	// Read the whole enc file
	result, err := io.ReadAll(encFile)
	require.NoError(t, err)

	// Save the file
	err = os.WriteFile("test/test.enc", result, 0644)
	require.NoError(t, err)

}

func TestDecrypt(t *testing.T) {
	file, err := os.Open("test/test.enc")
	require.NoError(t, err)
	defer file.Close()

	key := []byte("PPKpKqSMGfX43h2qJbP9cpkn886u9Y2D")
	decFile, err := Decrypt(key, file)
	require.NoError(t, err)

	result, err := io.ReadAll(decFile)
	require.NoError(t, err)

	err = os.WriteFile("test/test.dec", result, 0644)
	require.NoError(t, err)
}
