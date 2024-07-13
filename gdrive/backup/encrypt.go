package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func Encrypt(key []byte, file io.ReadCloser) (io.ReadCloser, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate a unique IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Create a new reader for the encrypted data
	encryptedReader := &cipher.StreamReader{
		S: cipher.NewCFBEncrypter(block, iv),
		R: file,
	}

	// Create a new reader that includes the IV at the beginning
	return &ivReader{
		iv:        iv,
		encrypted: encryptedReader,
	}, nil
}

func Decrypt(key []byte, file io.ReadCloser) (io.ReadCloser, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Read the first block of the file
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(file, iv); err != nil {
		return nil, err
	}

	// Create a new reader for the encrypted data
	encryptedReader := &cipher.StreamReader{
		S: cipher.NewCFBDecrypter(block, iv),
		R: file,
	}

	// Create a new reader that includes the IV at the beginning
	return &ivReader{
		iv:            iv,
		encrypted:     encryptedReader,
		EncryptedFile: true,
	}, nil
}

// ivReader is a reader that prepends the IV to the encrypted data
type ivReader struct {
	iv            []byte
	encrypted     io.Reader
	EncryptedFile bool
}

func (r *ivReader) Read(p []byte) (n int, err error) {
	// If we haven't written the IV yet, do it now
	if len(r.iv) > 0 {
		if r.EncryptedFile { // Just skip over this block
			r.iv = nil
			return 0, nil
		}

		n = copy(p, r.iv)
		r.iv = nil
		return n, nil
	}

	// Read the encrypted data
	return r.encrypted.Read(p)
}

func (r *ivReader) Close() error {
	return nil
	// return r.encrypted.Close()
}
