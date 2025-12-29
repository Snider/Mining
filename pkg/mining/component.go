package mining

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed component/*
var componentFS embed.FS

// GetComponentFS returns the embedded file system containing the web component.
// This allows the component to be served even when the package is used as a module.
func GetComponentFS() (http.FileSystem, error) {
	sub, err := fs.Sub(componentFS, "component")
	if err != nil {
		return nil, err
	}
	return http.FS(sub), nil
}
