package nextcloud

import (
	"encoding/json"
	"fmt"
	"io"
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
		log.Fatalf("%s", err)
	}
	client := gowebdav.NewClient(authDetails.Address, authDetails.Username, authDetails.Password)
	err = client.Connect()
	if err != nil {
		fmt.Println("Error connecting:", err)
	}

	return Client{client: client}, nil
}

func (c *Client) ListFiles() {
	files, err := c.client.ReadDir("/")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	fmt.Println("Files on WebDAV server:")
	for _, file := range files {
		fmt.Println(file.Name())
	}

}
