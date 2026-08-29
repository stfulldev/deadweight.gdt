package config

import (
	"bytes"
	"os"
	"path/filepath"
)

// Source identifies one selected configuration file and why it was selected.
type Source struct {
	Path     string
	Explicit bool
}

// Discover selects an explicit config or the project-local implicit default.
// A false present result is the normal missing-implicit case.
func Discover(projectRoot, explicitPath string) (source Source, present bool, err error) {
	if explicitPath != "" {
		path, pathErr := filepath.Abs(explicitPath)
		if pathErr != nil {
			return Source{}, false, configError(
				ReasonFilesystem,
				explicitPath,
				"",
				"cannot resolve explicit config path",
				pathErr,
			)
		}

		return discoverPath(filepath.Clean(path), true)
	}
	if projectRoot == "" {
		return Source{}, false, configError(
			ReasonValidation,
			"",
			"project_root",
			"project root is required for implicit config discovery",
			nil,
		)
	}

	path, pathErr := filepath.Abs(filepath.Join(projectRoot, DefaultFilename))
	if pathErr != nil {
		return Source{}, false, configError(
			ReasonFilesystem,
			filepath.Join(projectRoot, DefaultFilename),
			"",
			"cannot resolve implicit config path",
			pathErr,
		)
	}

	return discoverPath(filepath.Clean(path), false)
}

func discoverPath(path string, explicit bool) (Source, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if !explicit {
				return Source{}, false, nil
			}

			return Source{}, false, configError(
				ReasonMissingExplicit,
				path,
				"",
				"explicit config does not exist",
				err,
			)
		}

		return Source{}, false, configError(
			ReasonFilesystem,
			path,
			"",
			"cannot inspect config path",
			err,
		)
	}
	if !info.Mode().IsRegular() {
		return Source{}, false, configError(
			ReasonNotRegular,
			path,
			"",
			"config path is not a regular file",
			nil,
		)
	}

	return Source{Path: path, Explicit: explicit}, true, nil
}

// Read loads and decodes one previously discovered source.
func Read(source Source) (Config, error) {
	if source.Path == "" {
		return Config{}, configError(
			ReasonValidation,
			"",
			"source",
			"config source path is required",
			nil,
		)
	}

	data, err := os.ReadFile(source.Path)
	if err != nil {
		return Config{}, configError(
			ReasonFilesystem,
			source.Path,
			"",
			"cannot read config file",
			err,
		)
	}

	return Decode(bytes.NewReader(data), source.Path)
}

// Load discovers and reads configuration in one composition-friendly call.
// Source and present retain selection evidence even when reading fails.
func Load(projectRoot, explicitPath string) (Config, Source, bool, error) {
	source, present, err := Discover(projectRoot, explicitPath)
	if err != nil || !present {
		return Config{}, source, present, err
	}

	configuration, err := Read(source)
	if err != nil {
		return Config{}, source, true, err
	}

	return configuration, source, true, nil
}
