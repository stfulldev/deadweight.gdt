package project

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const projectFileName = "project.godot"

// StatFunc returns filesystem metadata for a host path.
type StatFunc func(name string) (fs.FileInfo, error)

// Finder discovers Godot project roots using filesystem metadata only.
type Finder struct {
	stat StatFunc
}

// NewFinder constructs a finder backed by the host filesystem.
func NewFinder() Finder {
	return NewFinderWithStat(os.Stat)
}

// NewFinderWithStat constructs a finder with an injected metadata operation.
func NewFinderWithStat(stat StatFunc) Finder {
	if stat == nil {
		stat = os.Stat
	}

	return Finder{stat: stat}
}

// Find discovers the project root for one scene input.
func (finder Finder) Find(request Request) (Root, error) {
	cwd, err := normalizeWorkingDirectory(request.WorkingDirectory)
	if err != nil {
		return Root{}, err
	}

	var explicit *Root
	if request.ExplicitProject != "" {
		root, findErr := finder.validateExplicitProject(request.ExplicitProject, cwd)
		if findErr != nil {
			return Root{}, findErr
		}
		explicit = &root
	}

	isResourcePath := strings.HasPrefix(request.SceneInput, "res://")
	var start string
	if isResourcePath {
		start = cwd
	} else {
		scene, findErr := finder.validateFilesystemScene(request.SceneInput, cwd)
		if findErr != nil {
			return Root{}, findErr
		}
		start = filepath.Dir(scene)
	}

	if explicit != nil {
		return *explicit, nil
	}
	if isResourcePath {
		if findErr := finder.validateWorkingDirectory(start); findErr != nil {
			return Root{}, findErr
		}
	}

	return finder.discoverFrom(start)
}

// FindContext discovers a project for a command that does not consume a scene.
func (finder Finder) FindContext(request ContextRequest) (Root, error) {
	cwd, err := normalizeWorkingDirectory(request.WorkingDirectory)
	if err != nil {
		return Root{}, err
	}

	if request.ExplicitProject != "" {
		return finder.validateExplicitProject(request.ExplicitProject, cwd)
	}
	if err := finder.validateWorkingDirectory(cwd); err != nil {
		return Root{}, err
	}

	return finder.discoverFrom(cwd)
}

func normalizeWorkingDirectory(raw string) (string, error) {
	if raw == "" || !filepath.IsAbs(raw) {
		return "", &Error{
			Reason: ReasonInvalidWorkingDirectory,
			Path:   raw,
			Detail: "must be an absolute path",
		}
	}

	return filepath.Clean(raw), nil
}

func resolveHostPath(raw, cwd string) string {
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw)
	}

	return filepath.Clean(filepath.Join(cwd, raw))
}

func (finder Finder) validateExplicitProject(raw, cwd string) (Root, error) {
	path := resolveHostPath(raw, cwd)
	info, err := finder.stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Root{}, invalidExplicitProject(raw, "path does not exist", err)
		}

		return Root{}, filesystemError(path, err)
	}

	if info.IsDir() {
		marker := filepath.Join(path, projectFileName)
		markerInfo, markerErr := finder.stat(marker)
		if markerErr != nil {
			if errors.Is(markerErr, fs.ErrNotExist) {
				return Root{}, invalidExplicitProject(raw, "directory does not contain a regular project.godot", markerErr)
			}

			return Root{}, filesystemError(marker, markerErr)
		}
		if !markerInfo.Mode().IsRegular() {
			return Root{}, invalidExplicitProject(raw, "project.godot is not a regular file", nil)
		}

		return newRoot(path), nil
	}

	if info.Mode().IsRegular() && filepath.Base(path) == projectFileName {
		return newRoot(filepath.Dir(path)), nil
	}

	return Root{}, invalidExplicitProject(raw, "must be a project directory or a regular project.godot file", nil)
}

func (finder Finder) validateFilesystemScene(raw, cwd string) (string, error) {
	path := resolveHostPath(raw, cwd)
	if filepath.Ext(path) != ".tscn" {
		return "", &Error{
			Reason: ReasonInvalidSceneInput,
			Path:   raw,
			Detail: "must have the case-sensitive .tscn extension",
		}
	}

	info, err := finder.stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", &Error{
				Reason: ReasonInvalidSceneInput,
				Path:   raw,
				Detail: "scene file does not exist",
				Err:    err,
			}
		}

		return "", filesystemError(path, err)
	}
	if !info.Mode().IsRegular() {
		return "", &Error{
			Reason: ReasonInvalidSceneInput,
			Path:   raw,
			Detail: "scene is not a regular file",
		}
	}

	return path, nil
}

func (finder Finder) validateWorkingDirectory(path string) error {
	info, err := finder.stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Error{
				Reason: ReasonInvalidWorkingDirectory,
				Path:   path,
				Detail: "directory does not exist",
				Err:    err,
			}
		}

		return filesystemError(path, err)
	}
	if !info.IsDir() {
		return &Error{
			Reason: ReasonInvalidWorkingDirectory,
			Path:   path,
			Detail: "must identify a directory",
		}
	}

	return nil
}

func (finder Finder) discoverFrom(start string) (Root, error) {
	start = filepath.Clean(start)
	for current := start; ; current = filepath.Dir(current) {
		marker := filepath.Join(current, projectFileName)
		info, err := finder.stat(marker)
		if err == nil && info.Mode().IsRegular() {
			return newRoot(current), nil
		}
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return Root{}, filesystemError(marker, err)
		}

		if parent := filepath.Dir(current); parent == current {
			break
		}
	}

	return Root{}, &Error{Reason: ReasonProjectNotFound, Path: start}
}

func newRoot(directory string) Root {
	directory = filepath.Clean(directory)
	return Root{
		Directory:   directory,
		ProjectFile: filepath.Join(directory, projectFileName),
	}
}

func invalidExplicitProject(path, detail string, err error) *Error {
	return &Error{
		Reason: ReasonInvalidExplicitProject,
		Path:   path,
		Detail: detail,
		Err:    err,
	}
}

func filesystemError(path string, err error) *Error {
	return &Error{Reason: ReasonFilesystem, Path: path, Err: err}
}
