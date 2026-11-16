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

// ReadTemplateFile reads a template file from the file system. The file,
// once its symbolic links are resolved, must be under the root, which is
// resolved too: a link under the root reads no file out of it. An error
// names the template, never a path of the host, since a render writes it.
func (l *LocalFileSystem) ReadTemplateFile(templatePath string) (string, error) {
	fullPath, err := l.FullPath(templatePath)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", NewFileSystemError(fmt.Sprintf("No such template '%s'", templatePath))
		}
		return "", NewFileSystemError(fmt.Sprintf("Failed to read template '%s'", templatePath))
	}
	root, err := filepath.EvalSymlinks(l.root)
	if err != nil {
		return "", NewFileSystemError(fmt.Sprintf("Failed to read template '%s'", templatePath))
	}
	if !under(resolved, root) {
		return "", NewFileSystemError(fmt.Sprintf("Illegal template path '%s'", templatePath))
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return "", NewFileSystemError(fmt.Sprintf("No such template '%s'", templatePath))
		}
		return "", NewFileSystemError(fmt.Sprintf("Failed to read template '%s'", templatePath))
	}
	return string(data), nil
}

// FullPath returns the full path to a template file, under the root.
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
		return "", NewFileSystemError(fmt.Sprintf("Illegal template path '%s'", templatePath))
	}
	absRoot, err := filepath.Abs(l.root)
	if err != nil {
		return "", NewFileSystemError(fmt.Sprintf("Illegal template path '%s'", templatePath))
	}
	if !under(absFullPath, absRoot) {
		return "", NewFileSystemError(fmt.Sprintf("Illegal template path '%s'", templatePath))
	}
	return absFullPath, nil
}

// under reports whether path is root or a path under it; a sibling whose
// name starts with the root's (root-other) is not.
func under(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
