package gdrive

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Client struct {
	client     *drive.Service
	baseFolder string
	Folders    map[string]string // a cached view of folder -> ID
}

func NewClient(authFlag, baseFolder string) (*Client, error) {
	b, err := os.ReadFile("../creds.json")
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

	return &Client{client: srv, baseFolder: baseFolder}, nil
}

func (c *Client) ListFiles() ([]*drive.File, error) {
	return c.listFiles(c.baseFolder)
}

func (c *Client) listFiles(baseFolder string) ([]*drive.File, error) {
	var allFiles []*drive.File
	pageToken := ""
	for {
		query := c.client.Files.List().Q("'" + baseFolder + "' in parents").Fields("files")
		if pageToken != "" {
			query = query.PageToken(pageToken)
		}
		r, err := query.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve files: %v", err)
		}
		allFiles = append(allFiles, r.Files...)
		// Go through all the folders
		for _, file := range r.Files {
			if file.MimeType == "application/vnd.google-apps.folder" {
				folderFiles, err := c.listFiles(file.Id)
				if err != nil {
					return nil, err
				}
				allFiles = append(allFiles, folderFiles...)
			}
		}

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

type File struct {
	Name   string
	Path   string
	Reader io.ReadCloser
}

func (c *Client) UploadFile(file File) error {
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

func (c *Client) GetFolderByID(folderID string) (*drive.File, error) {
	// Get the folder details
	folder, err := c.client.Files.Get(folderID).Fields("id,parents,name").Do()
	if err != nil {
		return nil, err
	}

	return folder, nil
}

func (client *Client) GetFullPath(parentID string) string {
	if parentID == "" {
		return ""
	}

	// Get the name of the parent folder
	parentFolder, err := client.GetFolderByID(parentID)
	if err != nil {
		log.Printf("Error getting parent folder: %v", err)
		return ""
	}

	var parentPath string
	if parentFolder.Id != client.baseFolder && len(parentFolder.Parents) > 0 {
		parentPath = client.GetFullPath(parentFolder.Parents[0])
	} else if parentFolder.Id == client.baseFolder {
		return parentPath
	}

	// Construct the full path
	return parentPath + "/" + parentFolder.Name
}
