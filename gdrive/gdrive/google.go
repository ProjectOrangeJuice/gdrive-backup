package gdrive

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ProjectOrangeJuice/gdrive-backup/gdrive/backup"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Client struct {
	client     *drive.Service
	baseFolder string
	Folders    map[string]string // a cached view of folder -> ID
}

func NewClient(authFlag string) (*Client, error) {
	b, err := os.ReadFile("../../creds.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file, %s", err)
	}

	if authFlag != "" {
		return nil, handleToken(authFlag, config)
	}
	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}
	ctx := context.Background()

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return &Client{client: srv}, nil
}

func (c *Client) ListFiles() ([]*drive.File, error) {
	var allFiles []*drive.File
	pageToken := ""
	for {
		query := c.client.Files.List().Q("'" + c.baseFolder + "' in parents").Fields("files(name, modifiedTime)")
		if pageToken != "" {
			query = query.PageToken(pageToken)
		}
		r, err := query.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve files: %v", err)
		}
		allFiles = append(allFiles, r.Files...)
		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}
	return allFiles, nil
}

func (c *Client) GetFolder(folderPath string) (string, error) {
	if id, ok := c.Folders[folderPath]; ok {
		return id, nil
	}

	// Split the path into individual folders
	folders := strings.Split(folderPath, "/")

	// Create the folders recursively
	var parentID string
	for i, folder := range folders {
		if i == 0 && folder == "" {
			continue // Skip empty first folder
		}
		folderID, err := c.createFolder(folder, parentID)
		if err != nil {
			return "", fmt.Errorf("error creating folder: %v", err)
		}
		parentID = folderID
		c.Folders[folderPath] = folderID // Cache the full path
	}

	return parentID, nil
}

func (c *Client) createFolder(folderName, parentID string) (string, error) {
	// Search for the folder
	r, err := c.client.Files.List().Q(fmt.Sprintf("'%s' in parents mimeType='application/vnd.google-apps.folder' and name='%s'", c.baseFolder, folderName)).
		Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("error listing files: %v", err)
	}

	var folderID string
	if len(r.Files) > 0 {
		folderID = r.Files[0].Id
		log.Printf("Folder '%s' already exists (ID: %s)\n", folderName, folderID)
	} else {
		if parentID == "" {
			parentID = c.baseFolder
		}
		// Create the folder
		folderMetadata := &drive.File{
			Name:     folderName,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{parentID},
		}
		folder, err := c.client.Files.Create(folderMetadata).Do()
		if err != nil {
			return "", fmt.Errorf("error creating folder: %v", err)
		}
		folderID = folder.Id
		log.Printf("Created folder '%s' (ID: %s)\n", folderName, folderID)
	}

	return folderID, nil
}

func (c *Client) UploadFile(file backup.File) error {
	defer file.Reader.Close()

	folderID, err := c.GetFolder(file.Path)
	if err != nil {
		return fmt.Errorf("unable to get folder: %v", err)
	}
	// Create Drive file metadata
	driveFile := &drive.File{
		Name:    file.Name, // Use the filename as the Drive file name
		Parents: []string{folderID},
	}

	// Upload the file
	_, err = c.client.Files.Create(driveFile).Media(file.Reader).Do()
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}
	log.Printf("Uploaded %s", file.Name)
	return nil
}
