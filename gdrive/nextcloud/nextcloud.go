package nextcloud

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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
	f, err := os.Open("../../nextcloud.json")
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

func NewClient() (Client, error) {
	authDetails, err := getAuth()
	if err != nil {
		return Client{}, err
	}
	client := gowebdav.NewClient(authDetails.Address, authDetails.Username, authDetails.Password)
	err = client.Connect()
	if err != nil {
		return Client{}, fmt.Errorf("error connecting: %s", err)
	}

	return Client{client: client}, nil
}

func (c *Client) ListFiles(dir string) ([]fs.FileInfo, error) {
	files, err := c.client.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory, %s", err)
	}

	fmt.Println("Files on WebDAV server:")
	for _, file := range files {
		fmt.Println(file.Name())
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
