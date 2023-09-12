package envi

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Envi struct {
	loadedVars map[string]string
}

// NewEnvi creates a new Envi instance.
func NewEnvi() *Envi {
	return &Envi{
		loadedVars: make(map[string]string),
	}
}

// FromMap loads the given key-value pairs and loads them into the local map.
func (envi *Envi) FromMap(vars map[string]string) {
	for key := range vars {
		envi.loadedVars[key] = vars[key]
	}
}

// LoadEnv loads the given keys from the environment variables.
func (envi *Envi) LoadEnv(vars ...string) {
	for _, key := range vars {
		envi.loadedVars[key] = os.Getenv(key)
	}
}

// LoadYAMLFilesFromEnvPaths loads yaml files from the paths in the given environment variables.
func (envi *Envi) LoadYAMLFilesFromEnvPaths(vars ...string) error {
	const errMessage = "failed to load yaml files from env paths: %w"
	for _, key := range vars {
		path := os.Getenv(key)

		if path == "" {
			return fmt.Errorf(errMessage, &ErrEnvVarNotFound{key})
		}

		if err := envi.LoadYAMLFile(path); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadYAMLFilesFromEnvPaths loads json files from the paths in the given environment variables.
func (envi *Envi) LoadJSONFilesFromEnvPaths(vars ...string) error {
	const errMessage = "failed to load json files from env paths: %w"

	for _, key := range vars {
		path := os.Getenv(key)

		if path == "" {
			return fmt.Errorf(errMessage, &ErrEnvVarNotFound{key})
		}

		if err := envi.LoadJSONFile(path); err != nil {
			return errors.Wrapf(err, "failed to read file '%s'", path)
		}
	}

	return nil
}

// LoadYAMLFilesFromEnvPaths loads the file content from the paths in the given environment variable
// to the value of the given key.
func (envi *Envi) LoadFileFromEnvPath(key string, envPath string) error {
	const errMessage = "failed to load file from env paths: %w"

	filePath := os.Getenv(envPath)

	if filePath == "" {
		return fmt.Errorf(errMessage, &ErrEnvVarNotFound{envPath})
	}

	if err := envi.LoadFile(key, filePath); err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadFile loads a string value under given key from a file.
func (envi *Envi) LoadFile(key, filePath string) error {
	blob, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read file '%s'", filePath)
	}

	envi.loadedVars[key] = string(blob)

	return nil
}

// LoadJSONFiles loads key-value pairs from one or more json files.
func (envi *Envi) LoadJSONFiles(paths ...string) error {
	for i := range paths {
		if err := envi.LoadJSONFile(paths[i]); err != nil {
			return errors.Wrapf(err, "failed to read json file '%s'", paths[i])
		}
	}

	return nil
}

// LoadJSONFile loads key-value pairs from a json files.
func (envi *Envi) LoadJSONFile(path string) error {
	blob, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read json file '%s'", path)
	}

	err = envi.LoadJSON(blob)
	if err != nil {
		return errors.Wrapf(err, "failed to load json file '%s'", path)
	}

	return nil
}

// LoadJSON loads key-value pairs from one or many json blobs.
func (envi *Envi) LoadJSON(blobs ...[]byte) error {
	for i := range blobs {
		var decoded map[string]string

		err := json.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal json")
		}

		for key := range decoded {
			envi.loadedVars[key] = decoded[key]
		}
	}

	return nil
}

// LoadYAMLFiles loads key-value pairs from one or more yaml files.
func (envi *Envi) LoadYAMLFiles(paths ...string) error {
	for i := range paths {
		if err := envi.LoadYAMLFile(paths[i]); err != nil {
			return errors.Wrapf(err, "failed to read yaml file '%s'", paths[i])
		}
	}

	return nil
}

// LoadYAMLFile loads key-value pairs from a yaml files.
func (envi *Envi) LoadYAMLFile(path string) error {
	blob, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "failed to read yaml file '%s'", path)
	}

	err = envi.LoadYAML(blob)
	if err != nil {
		return errors.Wrapf(err, "failed to load yaml file '%s'", path)
	}

	return nil
}

// LoadYAML loads key-value pairs from one or many yaml blobs.
func (envi *Envi) LoadYAML(blobs ...[]byte) error {
	for i := range blobs {
		var decoded map[string]string

		err := yaml.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal yaml")
		}

		for key := range decoded {
			envi.loadedVars[key] = decoded[key]
		}
	}

	return nil
}

// EnsureVars checks, if all given keys have a non-empty value.
func (envi *Envi) EnsureVars(requiredVars ...string) error {
	var missingVars []string

	for _, key := range requiredVars {
		if envi.loadedVars[key] == "" {
			missingVars = append(missingVars, key)
		}
	}

	if len(missingVars) > 0 {
		return &RequiredEnvVarsMissing{MissingVars: missingVars}
	}

	return nil
}

// ToEnv writes all key-value pairs to the environment.
func (envi *Envi) ToEnv() {
	for key := range envi.loadedVars {
		os.Setenv(key, envi.loadedVars[key])
	}
}

// ToMap returns a map, containing all key-value pairs.
func (envi *Envi) ToMap() map[string]string {
	return envi.loadedVars
}
