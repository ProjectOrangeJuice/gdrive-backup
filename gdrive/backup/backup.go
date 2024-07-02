package backup

import "io"

type File struct {
	Name   string
	Path   string
	Reader io.ReadCloser
}

type Config struct {
	Directories []string
}
