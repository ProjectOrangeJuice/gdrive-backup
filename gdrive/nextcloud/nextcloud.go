package nextcloud

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/studio-b12/gowebdav"
)

type auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Address  string `json:"address"`
}

type Client struct {
	client *gowebdav.Client
}

func getAuth() (auth, error) {
	var a auth
	f, err := os.Open("../nextcloud.json")
	if err != nil {
		return a, fmt.Errorf("could not read nextcloud.json, %s", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return a, fmt.Errorf("could not read nextcloud.json, %s", err)
	}

	err = json.Unmarshal(b, &a)
	if err != nil {
		return a, fmt.Errorf("could not read nextcloud.json, %s", err)
	}

	return a, nil
}

func NewClient() (*Client, error) {
	authDetails, err := getAuth()
	if err != nil {
		return nil, err
	}
	client := gowebdav.NewClient(authDetails.Address, authDetails.Username, authDetails.Password)
	err = client.Connect()
	if err != nil {
		return nil, fmt.Errorf("error connecting: %s", err)
	}

	return &Client{client: client}, nil
}

func (c *Client) ListFiles(dir string) ([]ExtraFileInfo, error) {
	files, err := c.client.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory, %s", err)
	}
	cov := make([]ExtraFileInfo, len(files))
	for i, file := range files {
		// If dir doesn't end in a slash, add it
		if dir[len(dir)-1] != '/' {
			dir += "/"
		}
		cov[i] = ExtraFileInfo{FileInfo: file, Path: dir + file.Name()}
	}

	return cov, nil
}

type ExtraFileInfo struct {
	fs.FileInfo
	Path string
}

func (c *Client) ListAllFiles(dir string) ([]ExtraFileInfo, error) {
	log.Printf("Looking at %s", dir)
	files, err := c.ListFiles(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			extraFiles, err := c.ListAllFiles(file.Path)
			if err != nil {
				return nil, err
			}
			files = append(files, extraFiles...)
		}
	}
	return files, nil
}

func (c *Client) DownloadFile(path string) (io.ReadCloser, error) {
	reader, err := c.client.ReadStream(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file, %s", err)
	}

	return reader, nil
}
