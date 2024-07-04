package gdrive

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Client struct {
	client     *drive.Service
	baseFolder string
	folderLock sync.Mutex        // so we can cache the folders without conflict
	Folders    map[string]string // a cached view of folder -> ID
	FolderIDs  map[string]string // a cached view of folderID -> Folder path
}

const Scope = drive.DriveFileScope

func NewClient(authFlag, baseFolder string) (*Client, error) {
	b, err := os.ReadFile("../creds.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, Scope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file, %s", err)
	}

	if authFlag != "" {
		err := handleToken(authFlag, config)
		if err != nil {
			return nil, err
		}
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
	folders := make(map[string]string)
	folderIDs := make(map[string]string)
	return &Client{client: srv,
		baseFolder: baseFolder, Folders: folders,
		FolderIDs:  folderIDs,
		folderLock: sync.Mutex{}}, nil
}

func (c *Client) ListFiles() ([]*drive.File, error) {
	return c.listFiles(c.baseFolder)
}

func (c *Client) listFiles(baseFolder string) ([]*drive.File, error) {
	var allFiles []*drive.File
	pageToken := ""
	for {
		query := c.client.Files.List().Q("'" + baseFolder + "' in parents and trashed=false").Fields("files")
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
	c.folderLock.Lock()
	defer c.folderLock.Unlock()

	// Split the path into individual folders
	folders := strings.Split(folderPath, "/")

	// Create the folders recursively
	var parentID string
	for _, folder := range folders {
		if folder == "" {
			continue
		}
		folderID, err := c.createFolder(folder, parentID)
		if err != nil {
			return "", fmt.Errorf("error creating folder: %v", err)
		}
		parentID = folderID
	}
	c.Folders[folderPath] = parentID // Cache the full path

	return parentID, nil
}

func (c *Client) createFolder(folderName, parentID string) (string, error) {
	if parentID == "" {
		parentID = c.baseFolder
	}
	// Search for the folder
	r, err := c.client.Files.List().Q(fmt.Sprintf("'%s' in parents and mimeType='application/vnd.google-apps.folder' and name='%s' and trashed=false", parentID, folderName)).
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
	Name         string
	Path         string
	ModifiedTime time.Time
	Reader       io.ReadCloser
}

func (c *Client) UploadFile(file File) error {
	defer file.Reader.Close()

	// file path without the file name
	fp := strings.TrimSuffix(file.Path, file.Name)
	folderID, err := c.GetFolder(fp)
	if err != nil {
		return fmt.Errorf("unable to get folder: %v", err)
	}
	// get existing file
	existing, err := c.GetFile(file.Name, folderID)
	if err != nil {
		return fmt.Errorf("unable to get existing file: %v", err)
	}

	// Create Drive file metadata
	driveFile := &drive.File{
		Name:         file.Name, // Use the filename as the Drive file name
		Parents:      []string{folderID},
		ModifiedTime: file.ModifiedTime.Format(time.RFC3339),
	}

	// Upload the file
	_, err = c.client.Files.Create(driveFile).Media(file.Reader).Do()
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}
	log.Printf("Uploaded %s", file.Name)
	if existing != nil {
		c.DeleteFile(existing.Id)
	}
	return nil
}

func (c *Client) DeleteFile(fileID string) error {
	err := c.client.Files.Delete(fileID).Do()
	if err != nil {
		return fmt.Errorf("error deleting file: %v", err)
	}
	log.Printf("Deleted %s", fileID)
	return nil
}

func (c *Client) GetFolderByID(folderID string) (*drive.File, error) {
	// Get the folder details
	folder, err := c.client.Files.Get(folderID).Fields("id,parents,name").Do()
	if err != nil {
		return nil, fmt.Errorf("tried to get the folder [%s] but got an error, %s", folderID, err)
	}

	return folder, nil
}

func (c *Client) GetFile(fileName, parentFolderID string) (*drive.File, error) {
	r, err := c.client.Files.List().Q(fmt.Sprintf("'%s' in parents and name='%s' and trashed=false", parentFolderID, fileName)).
		Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		return nil, fmt.Errorf("error get file %s: %v", fileName, err)
	}

	if len(r.Files) > 1 {
		return nil, fmt.Errorf("matched more than one file when looking for %s in %s", fileName, parentFolderID)
	}
	if len(r.Files) == 0 {
		return nil, nil
	}

	return r.Files[0], nil
}

func (client *Client) GetFullPath(parentID string) (string, error) {
	if parentID == "" || parentID == client.baseFolder {
		return "", nil
	}

	if _, ok := client.FolderIDs[parentID]; ok {
		return client.FolderIDs[parentID], nil
	}

	// Get the name of the parent folder
	parentFolder, err := client.GetFolderByID(parentID)
	if err != nil {

		return "", fmt.Errorf("error getting parent folder: %v", err)
	}

	var parentPath string
	if parentFolder.Id != client.baseFolder && len(parentFolder.Parents) > 0 {
		parentPath, err = client.GetFullPath(parentFolder.Parents[0])
		if err != nil {
			return "", fmt.Errorf("error getting parent path for %s: %v", parentFolder.Name, err)
		}
	} else if parentFolder.Id == client.baseFolder {
		return parentPath, nil
	}

	// Construct the full path
	fullPath := parentPath + "/" + parentFolder.Name
	client.FolderIDs[parentID] = fullPath
	return fullPath, nil
}
