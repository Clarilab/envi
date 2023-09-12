package envi

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Envi is a config loader to load all sorts of configuration files.
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
			return fmt.Errorf(errMessage, &EnvVarNotFoundError{key})
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
			return fmt.Errorf(errMessage, &EnvVarNotFoundError{key})
		}

		if err := envi.LoadJSONFile(path); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadYAMLFilesFromEnvPaths loads the file content from the path in the given environment variable
// to the value of the given key.
func (envi *Envi) LoadFileFromEnvPath(key string, envPath string) error {
	const errMessage = "failed to load file from env paths: %w"

	filePath := os.Getenv(envPath)

	if filePath == "" {
		return fmt.Errorf(errMessage, &EnvVarNotFoundError{envPath})
	}

	if err := envi.LoadFile(key, filePath); err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadFile loads a string value under given key from a file.
func (envi *Envi) LoadFile(key, filePath string) error {
	const errMessage = "failed to load file: %w"

	blob, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{filePath})
	}

	envi.loadedVars[key] = string(blob)

	return nil
}

// LoadJSONFiles loads key-value pairs from one or more json files.
func (envi *Envi) LoadJSONFiles(paths ...string) error {
	const errMessage = "failed to load json files: %w"

	for i := range paths {
		if err := envi.LoadJSONFile(paths[i]); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadJSONFile loads key-value pairs from a json file.
func (envi *Envi) LoadJSONFile(path string) error {
	const errMessage = "failed to load json file: %w"

	blob, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{path})
	}

	err = envi.LoadJSON(blob)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadJSON loads key-value pairs from one or many json blobs.
func (envi *Envi) LoadJSON(blobs ...[]byte) error {
	const errMessage = "failed to load json: %w"

	for i := range blobs {
		var decoded map[string]string

		err := json.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return fmt.Errorf(errMessage, err)
		}

		for key := range decoded {
			envi.loadedVars[key] = decoded[key]
		}
	}

	return nil
}

// LoadYAMLFiles loads key-value pairs from one or more yaml files.
func (envi *Envi) LoadYAMLFiles(paths ...string) error {
	const errMessage = "failed to load yaml files: %w"

	for i := range paths {
		if err := envi.LoadYAMLFile(paths[i]); err != nil {
			return fmt.Errorf(errMessage, err)
		}
	}

	return nil
}

// LoadYAMLFile loads key-value pairs from a yaml file.
func (envi *Envi) LoadYAMLFile(path string) error {
	const errMessage = "failed to load yaml file: %w"

	blob, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf(errMessage, &FailedToReadFileError{path})
	}

	err = envi.LoadYAML(blob)
	if err != nil {
		return fmt.Errorf(errMessage, err)
	}

	return nil
}

// LoadYAML loads key-value pairs from one or many yaml blobs.
func (envi *Envi) LoadYAML(blobs ...[]byte) error {
	const errMessage = "failed to load yaml file: %w"

	for i := range blobs {
		var decoded map[string]string

		err := yaml.Unmarshal(blobs[i], &decoded)
		if err != nil {
			return fmt.Errorf(errMessage, err)
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
