package liquid

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// TemplateNameRegex validates template names (letters, numbers, underscore, slash)
	TemplateNameRegex = regexp.MustCompile(`^[^./][a-zA-Z0-9_/]+$`)
)

// FileSystem is an interface for retrieving template files.
type FileSystem interface {
	ReadTemplateFile(templatePath string) (string, error)
}

// BlankFileSystem is a file system that doesn't allow includes.
type BlankFileSystem struct{}

// ReadTemplateFile always returns an error for BlankFileSystem.
func (b *BlankFileSystem) ReadTemplateFile(_ string) (string, error) {
	return "", NewFileSystemError("This liquid context does not allow includes.")
}

// LocalFileSystem retrieves template files from the local file system.
// Template files are named with an underscore prefix and .liquid extension,
// similar to Rails partials.
type LocalFileSystem struct {
	root    string
	pattern string
}

// NewLocalFileSystem creates a new LocalFileSystem with the given root directory.
// The pattern defaults to "_%s.liquid" if not provided.
func NewLocalFileSystem(root, pattern string) *LocalFileSystem {
	if pattern == "" {
		pattern = "_%s.liquid"
	}
	return &LocalFileSystem{
		root:    root,
		pattern: pattern,
	}
}

// ReadTemplateFile reads a template file from the file system.
func (l *LocalFileSystem) ReadTemplateFile(templatePath string) (string, error) {
	fullPath, err := l.FullPath(templatePath)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", NewFileSystemError(fmt.Sprintf("No such template '%s'", templatePath))
		}
		return "", NewFileSystemError(fmt.Sprintf("Failed to read template '%s': %v", templatePath, err))
	}

	return string(data), nil
}

// FullPath returns the full path to a template file.
func (l *LocalFileSystem) FullPath(templatePath string) (string, error) {
	if !TemplateNameRegex.MatchString(templatePath) {
		return "", NewFileSystemError(fmt.Sprintf("Illegal template name '%s'", templatePath))
	}

	var fullPath string
	if strings.Contains(templatePath, "/") {
		dir := filepath.Dir(templatePath)
		base := filepath.Base(templatePath)
		fullPath = filepath.Join(l.root, dir, fmt.Sprintf(l.pattern, base))
	} else {
		fullPath = filepath.Join(l.root, fmt.Sprintf(l.pattern, templatePath))
	}

	// Security check: ensure the resolved path is within the root directory
	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", NewFileSystemError(fmt.Sprintf("Failed to resolve path: %v", err))
	}

	absRoot, err := filepath.Abs(l.root)
	if err != nil {
		return "", NewFileSystemError(fmt.Sprintf("Failed to resolve root: %v", err))
	}

	if !strings.HasPrefix(absFullPath, absRoot) {
		return "", NewFileSystemError(fmt.Sprintf("Illegal template path '%s'", absFullPath))
	}

	return fullPath, nil
}
